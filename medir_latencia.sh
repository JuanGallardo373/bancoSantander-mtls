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
            -CAfile CABancoCentral/cacert.pem \
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
echo "================================================================"
echo "BENCHMARK PKI DE UN NIVEL - OpenSSL s_time (10 corridas de 5s)"
echo "================================================================"
echo ""
echo "Asegúrate que:"
echo "   1. El servidor esta corriendo: go run main.go"
echo "   2. El servidor OCSP esta en localhost:2560"
echo ""

# Archivos temporales para almacenar estadísticas
stats_pki_one_level="/tmp/stats_pki_one_level.txt"

# Ejecutar benchmark para PKI de un nivel sin OCSP
echo "FASE 1: Midiendo latencia con PKI de un nivel..."
run_benchmark "PKI UN NIVEL SIN OCSP" "$stats_pki_one_level"

# ============================================================
# EXTRACCION DE ESTADISTICAS
# ============================================================

# Leer estadísticas desde archivos temporales
stats_pki_one_level=$(cat "$stats_pki_one_level")

# Extraer valores
pki_one_level_mean=$(echo "$stats_pki_one_level" | cut -d'|' -f1)
pki_one_level_std=$(echo "$stats_pki_one_level" | cut -d'|' -f2)
pki_one_level_ci_low=$(echo "$stats_pki_one_level" | cut -d'|' -f3)
pki_one_level_ci_up=$(echo "$stats_pki_one_level" | cut -d'|' -f4)

# Extraer todas las corridas para incluirlas en el reporte
pki_one_level_runs=$(echo "$stats_pki_one_level" | cut -d'|' -f5- | sed 's/|$//')

# ============================================================
# COMPARACION Y ANALISIS FINAL
# ============================================================

echo ""
echo "=========================================================="
echo "RESUMEN FINAL"
echo "=========================================================="
echo ""
echo "PKI DE UN NIVEL SIN OCSP (baseline)"
echo "=========================================================="
echo "Media:               $pki_one_level_mean ms/handshake"
echo "Desv. Estandar:      $pki_one_level_std ms"
echo "IC 95%:              [$pki_one_level_ci_low - $pki_one_level_ci_up] ms"
echo ""

# ============================================================
# GENERAR REPORTE EN MARKDOWN
# ============================================================

timestamp=$(date +%Y%m%d_%H%M%S)
mkdir -p data
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
    echo "- **Duración por corrida:** 5 segundos"
    echo "- **Protocolo:** TLS 1.3"
    echo "- **Autenticación:** mTLS con PKI de un nivel"  
    echo ""
    echo "---"
    echo ""
    echo "## Resultados: PKI de un nivel SIN OCSP (Baseline)"
    echo ""
    echo "| Métrica | Valor |"
    echo "|---------|-------|"
    echo "| **Latencia promedio** | $pki_one_level_mean ms/handshake |"
    echo "| **Desviación estándar** | $pki_one_level_std ms |"
    echo "| **IC 95% (inferior)** | $pki_one_level_ci_low ms |"
    echo "| **IC 95% (superior)** | $pki_one_level_ci_up ms |"
    echo "| **Rango de confianza** | [$pki_one_level_ci_low - $pki_one_level_ci_up] ms |"
    echo ""
    echo "### Detalle de las 10 corridas"
    echo ""
    echo "| Corrida | Latencia (ms) |"
    echo "|---------|---------------|"
    
    i=1
    IFS='|' read -ra runs_array <<< "$pki_one_level_runs"
    for latency in "${runs_array[@]}"; do
        if [ ! -z "$latency" ]; then
            echo "| Run $i | $latency |"
            i=$((i + 1))
        fi
    done
    
    echo "---"
    echo ""
    echo "## Comparativa Visual"
    echo ""
    echo "### Latencia Promedio"
    echo ""
    echo "\`\`\`"
    echo "PKI de un nivel sin OCSP:    $pki_one_level_mean ms"
    echo ""
    echo "### Rango de Confianza (95%)"
    echo ""
    echo "**PKI de un nivel sin OCSP:**    [$pki_one_level_ci_low - $pki_one_level_ci_up] ms"
    echo ""
    echo "---"
    echo ""
    echo "## Conclusiones"
    echo ""
    echo "- Se realizaron **10 corridas** de **5 segundos**"
    echo "- Se calcularon intervalos de confianza del 95% usando distribución t-Student"
    echo "- El análisis incluye desviación estándar y error estándar"
    echo ""
    
} > "$report_file"

echo "✅ Reporte guardado en: **$report_file**"
echo ""

# Limpiar archivos temporales
rm -f "$stats_pki_one_level"

echo "=========================================================="
echo "Benchmark completado exitosamente"
echo "=========================================================="
