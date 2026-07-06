#!/bin/bash

# Script para medir overhead OCSP con estadistica rigurosa
# Usa openssl s_time para 10 corridas

calculate_stats() {
    local -n arr=$1
    local n=${#arr[@]}
    
    if [ $n -eq 0 ]; then
        echo "0|0|0|0"
        return
    fi
    
    # Crear string con todos los valores separados por espacios
    local values_str=""
    for val in "${arr[@]}"; do
        values_str="$values_str $val"
    done
    
    # Calcular todo con awk
    echo "$values_str" | awk -v n=$n '
    BEGIN {
        n_vals = 0
        sum = 0
        sum_sq = 0
    }
    {
        for (i = 1; i <= NF; i++) {
            val = $i
            sum += val
            n_vals++
        }
    }
    END {
        if (n_vals > 0) {
            mean = sum / n_vals
            
            # Recalcular para desviacion estandar
            n_vals = 0
            sum = 0
            for (i = 1; i <= NF; i++) {
                val = $i
                diff = val - mean
                sum_sq += diff * diff
                n_vals++
            }
            
            if (n_vals > 1) {
                variance = sum_sq / (n_vals - 1)
            } else {
                variance = sum_sq / n_vals
            }
            
            std_dev = sqrt(variance)
            std_error = std_dev / sqrt(n_vals)
            
            # t-value para n=10 es ~2.262
            t_value = 2.262
            margin = t_value * std_error
            ci_lower = mean - margin
            ci_upper = mean + margin
            
            printf "%.4f|%.4f|%.4f|%.4f", mean, std_dev, ci_lower, ci_upper
        } else {
            printf "0|0|0|0"
        }
    }
    '
}

run_benchmark() {
    local scenario="$1"
    local stats_file="$2"  # Archivo donde guardar las estadísticas
    local num_runs=10
    local test_duration=3
    
    echo ""
    echo "=========================================================="
    echo "Benchmark: $scenario"
    echo "=========================================================="
    echo "Ejecutando $num_runs corridas de $test_duration segundos cada una..."
    echo ""
    
    declare -a latencies
    local run=1
    
    while [ $run -le $num_runs ]; do
        echo -n "Corrida $run/$num_runs: "
        
        # Ejecutar s_time y capturar output
        local output=$(openssl s_time \
            -connect localhost:8443 \
            -cert cliente-bancobbva/certs/bbva.crt \
            -key cliente-bancobbva/certs/bbva.key \
            -CAfile CAIntermediaBANELCO/bundle.crt \
            -time $test_duration \
            -new \
            -tls1_3 2>&1)
        
        # Extraer "X connections in Y real seconds"
        local conn_line=$(echo "$output" | grep "connections in.*real seconds")
        
        if [ -z "$conn_line" ]; then
            echo "Error: No se pudo extraer datos"
            run=$((run + 1))
            continue
        fi
        
        # Parsear: "1234 connections in 3.05 real seconds"
        local num_connections=$(echo "$conn_line" | awk '{print $1}')
        local real_seconds=$(echo "$conn_line" | awk '{print $4}')
        
        # Validar que sean numeros validos
        if ! [[ "$num_connections" =~ ^[0-9]+$ ]] || ! [[ "$real_seconds" =~ ^[0-9.]+$ ]]; then
            echo "Error: Datos invalidos"
            run=$((run + 1))
            continue
        fi
        
        # Calcular latencia promedio por handshake en ms con awk
        local latency_ms=$(echo "$real_seconds $num_connections" | awk '{printf "%.4f", ($1 * 1000) / $2}')
        
        latencies+=("$latency_ms")
        
        echo "$num_connections conexiones en ${real_seconds}s = ${latency_ms} ms/handshake"
        
        run=$((run + 1))
    done
    
    # Calcular estadisticas
    echo ""
    echo "Analisis estadistico:"
    echo "=========================================================="
    
    local stats=$(calculate_stats latencies)
    local mean=$(echo "$stats" | cut -d'|' -f1)
    local std_dev=$(echo "$stats" | cut -d'|' -f2)
    local ci_lower=$(echo "$stats" | cut -d'|' -f3)
    local ci_upper=$(echo "$stats" | cut -d'|' -f4)
    
    echo "Latencia promedio:        $mean ms/handshake"
    echo "Desviacion estandar:      $std_dev ms"
    echo "IC 95% (inferior):        $ci_lower ms"
    echo "IC 95% (superior):        $ci_upper ms"
    echo "Rango de confianza:       [$ci_lower - $ci_upper] ms"
    echo ""
 
    # Guardar estadísticas en archivo con formato pipe-delimited para recuperarlas después
    # Formato: mean|std_dev|ci_lower|ci_upper|run1|run2|...|run10
    local runs_str=""
    for latency in "${latencies[@]}"; do
        runs_str="${runs_str}${latency}|"
    done
    echo "${mean}|${std_dev}|${ci_lower}|${ci_upper}|${runs_str}" > "$stats_file"
}

# ============================================================
# EJECUCION PRINCIPAL
# ============================================================

echo ""
echo "=========================================================="
echo "BENCHMARK OCSP - OpenSSL s_time (10 corridas de 3s)"
echo "=========================================================="
echo ""
echo "Asegúrate que:"
echo "   1. El servidor esta corriendo: go run main.go"
echo "   2. El servidor OCSP esta en localhost:2560"
echo ""

# Archivos temporales para almacenar estadísticas
stats_with_ocsp="/tmp/stats_with_ocsp.txt"
stats_without_ocsp="/tmp/stats_without_ocsp.txt"

# Ejecutar CON OCSP
echo "FASE 1: Midiendo CON OCSP"
run_benchmark "CON OCSP HABILITADO" "$stats_with_ocsp"

echo ""
echo "PAUSA: Ahora desactiva OCSP"
echo "=========================================================="
echo "Pasos:"
echo "  1. Abre servidor-banco/main.go"
echo "  2. Comenta la seccion VerifyPeerCertificate: func(..."
echo "  3. Guarda y recompila: go build"
echo "  4. Reinicia el servidor: go run main.go"
echo ""
read -p "Presiona ENTER cuando este listo..."

# Ejecutar SIN OCSP
echo ""
echo "FASE 2: Midiendo SIN OCSP"
run_benchmark "SIN OCSP (baseline)" "$stats_without_ocsp"

# ============================================================
# EXTRACCION DE ESTADISTICAS
# ============================================================

# Leer estadísticas desde archivos temporales
stats_with=$(cat "$stats_with_ocsp")
stats_without=$(cat "$stats_without_ocsp")

# Extraer valores CON OCSP
with_mean=$(echo "$stats_with" | cut -d'|' -f1)
with_std=$(echo "$stats_with" | cut -d'|' -f2)
with_ci_low=$(echo "$stats_with" | cut -d'|' -f3)
with_ci_up=$(echo "$stats_with" | cut -d'|' -f4)

# Extraer valores SIN OCSP
without_mean=$(echo "$stats_without" | cut -d'|' -f1)
without_std=$(echo "$stats_without" | cut -d'|' -f2)
without_ci_low=$(echo "$stats_without" | cut -d'|' -f3)
without_ci_up=$(echo "$stats_without" | cut -d'|' -f4)

# Extraer todas las corridas para incluirlas en el reporte
with_runs=$(echo "$stats_with" | cut -d'|' -f5- | sed 's/|$//')
without_runs=$(echo "$stats_without" | cut -d'|' -f5- | sed 's/|$//')

# ============================================================
# COMPARACION Y ANALISIS FINAL
# ============================================================

echo ""
echo "=========================================================="
echo "RESUMEN COMPARATIVO FINAL"
echo "=========================================================="
echo ""

echo "CON OCSP"
echo "=========================================================="
echo "Media:               $with_mean ms/handshake"
echo "Desv. Estandar:      $with_std ms"
echo "IC 95%:              [$with_ci_low - $with_ci_up] ms"
echo ""

echo "SIN OCSP (baseline)"
echo "=========================================================="
echo "Media:               $without_mean ms/handshake"
echo "Desv. Estandar:      $without_std ms"
echo "IC 95%:              [$without_ci_low - $without_ci_up] ms"
echo ""

# Calcular overhead con awk
overhead_info=$(echo "$with_mean $without_mean" | awk '{
    abs_diff = $1 - $2
    if ($2 > 0) {
        pct_diff = (($1 - $2) / $2) * 100
    } else {
        pct_diff = 0
    }
    printf "%.4f|%.2f", abs_diff, pct_diff
}')

overhead_abs=$(echo "$overhead_info" | cut -d'|' -f1)
overhead_pct=$(echo "$overhead_info" | cut -d'|' -f2)

echo "OVERHEAD OCSP"
echo "=========================================================="
echo "Diferencia absoluta:  $overhead_abs ms"
echo "Diferencia porcentual: $overhead_pct %"
echo ""
echo "Interpretacion:"

# Comparaciones con awk para evitar problemas
result=$(echo "$overhead_pct" | awk '{
    if ($1 > 5) {
        print "ADVERTENCIA"
    } else if ($1 > 0) {
        print "INFO"
    } else {
        print "OK"
    }
}')

if [ "$result" = "ADVERTENCIA" ]; then
    echo "   ADVERTENCIA: OCSP anade ~${overhead_abs}ms por handshake"
    echo "   (~${overhead_pct}% mas lento)"
elif [ "$result" = "INFO" ]; then
    echo "   INFO: OCSP tiene impacto minimo (~${overhead_pct}%)"
else
    echo "   OK: OCSP esta dentro del margen de error"
fi

echo ""

# ============================================================
# GENERAR REPORTE EN MARKDOWN
# ============================================================

timestamp=$(date +%Y%m%d_%H%M%S)
report_file="data/pki_comparison_results_${timestamp}.md"

{
    echo "# Reporte Benchmark OCSP"
    echo ""
    echo "**Fecha:** $(date '+%d/%m/%Y %H:%M:%S')"
    echo ""
    echo "## Configuración del Benchmark"
    echo ""
    echo "- **Herramienta:** OpenSSL s_time"
    echo "- **Corridas por escenario:** 10"
    echo "- **Duración por corrida:** 3 segundos"
    echo "- **Protocolo:** TLS 1.3"
    echo "- **Autenticación:** mTLS con PKI jerárquica"
    echo ""
    echo "---"
    echo ""
    echo "## Resultados: CON OCSP HABILITADO"
    echo ""
    echo "| Métrica | Valor |"
    echo "|---------|-------|"
    echo "| **Latencia promedio** | $with_mean ms/handshake |"
    echo "| **Desviación estándar** | $with_std ms |"
    echo "| **IC 95% (inferior)** | $with_ci_low ms |"
    echo "| **IC 95% (superior)** | $with_ci_up ms |"
    echo "| **Rango de confianza** | [$with_ci_low - $with_ci_up] ms |"
    echo ""
    echo "### Detalle de las 10 corridas"
    echo ""
    echo "| Corrida | Latencia (ms) |"
    echo "|---------|---------------|"
    
    i=1
    IFS='|' read -ra runs_array <<< "$with_runs"
    for latency in "${runs_array[@]}"; do
        if [ ! -z "$latency" ]; then
            echo "| Run $i | $latency |"
            i=$((i + 1))
        fi
    done
    
    echo ""
    echo "---"
    echo ""
    echo "## Resultados: SIN OCSP (Baseline)"
    echo ""
    echo "| Métrica | Valor |"
    echo "|---------|-------|"
    echo "| **Latencia promedio** | $without_mean ms/handshake |"
    echo "| **Desviación estándar** | $without_std ms |"
    echo "| **IC 95% (inferior)** | $without_ci_low ms |"
    echo "| **IC 95% (superior)** | $without_ci_up ms |"
    echo "| **Rango de confianza** | [$without_ci_low - $without_ci_up] ms |"
    echo ""
    echo "### Detalle de las 10 corridas"
    echo ""
    echo "| Corrida | Latencia (ms) |"
    echo "|---------|---------------|"
    
    i=1
    IFS='|' read -ra runs_array <<< "$without_runs"
    for latency in "${runs_array[@]}"; do
        if [ ! -z "$latency" ]; then
            echo "| Run $i | $latency |"
            i=$((i + 1))
        fi
    done
    
    echo ""
    echo "---"
    echo ""
    echo "## Análisis de Overhead OCSP"
    echo ""
    echo "| Métrica | Valor |"
    echo "|---------|-------|"
    echo "| **Diferencia absoluta** | $overhead_abs ms |"
    echo "| **Diferencia porcentual** | $overhead_pct % |"
    echo ""
    echo "### Interpretación"
    echo ""
    
    if [ "$result" = "ADVERTENCIA" ]; then
        echo "⚠️ **ADVERTENCIA:** OCSP añade aproximadamente **${overhead_abs}ms** por handshake"
        echo ""
        echo "Esto representa un incremento del **${overhead_pct}%** en la latencia."
        echo ""
        echo "**Recomendaciones:**"
        echo "- Considerar implementar OCSP stapling para reducir el impacto"
        echo "- Evaluar si la validación de revocación es crítica para tu caso de uso"
        echo "- Implementar caching de respuestas OCSP"
    elif [ "$result" = "INFO" ]; then
        echo "ℹ️ **INFO:** OCSP tiene un impacto **mínimo** (~${overhead_pct}%)"
        echo ""
        echo "El overhead es aceptable y puede ser ignorado en la mayoría de casos."
    else
        echo "✅ **OK:** OCSP está dentro del **margen de error**"
        echo ""
        echo "La diferencia observada puede deberse a variabilidad natural del sistema."
        echo "No hay evidencia de impacto significativo por OCSP."
    fi
    
    echo ""
    echo "---"
    echo ""
    echo "## Comparativa Visual"
    echo ""
    echo "### Latencia Promedio"
    echo ""
    echo "\`\`\`"
    echo "Con OCSP:    $with_mean ms"
    echo "Sin OCSP:    $without_mean ms"
    echo "Overhead:    $overhead_abs ms ($overhead_pct%)"
    echo "\`\`\`"
    echo ""
    echo "### Rango de Confianza (95%)"
    echo ""
    echo "**Con OCSP:**    [$with_ci_low - $with_ci_up] ms"
    echo "**Sin OCSP:**    [$without_ci_low - $without_ci_up] ms"
    echo ""
    echo "---"
    echo ""
    echo "## Conclusiones"
    echo ""
    echo "- Se realizaron **10 corridas** de **3 segundos** cada una por escenario"
    echo "- Se calcularon intervalos de confianza del 95% usando distribución t-Student"
    echo "- El análisis incluye desviación estándar y error estándar"
    echo ""
    
} > "$report_file"

echo "✅ Reporte guardado en: **$report_file**"
echo ""

# Limpiar archivos temporales
rm -f "$stats_with_ocsp" "$stats_without_ocsp"

echo "=========================================================="
echo "Benchmark completado exitosamente"
echo "=========================================================="