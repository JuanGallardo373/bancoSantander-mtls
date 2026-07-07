# Reporte Benchmark OCSP en PKI de 2 niveles

**Fecha:** 07/07/2026 12:00:29

## Configuración del Benchmark

- **Herramienta:** OpenSSL s_time
- **Corridas por escenario:** 10
- **Duración por corrida:** 3 segundos
- **Protocolo:** TLS 1.3
- **Autenticación:** mTLS con PKI jerárquica

---

## Resultados: CON OCSP HABILITADO

| Métrica | Valor |
|---------|-------|
| **Latencia promedio** | 10,4400 ms/handshake |
| **Desviación estándar** | 2,1127 ms |
| **IC 95% (inferior)** | 8,9288 ms |
| **IC 95% (superior)** | 11,9513 ms |
| **Rango de confianza** | [8,9288 - 11,9513] ms |

### Detalle de las 10 corridas

| Corrida | Latencia (ms) |
|---------|---------------|
| Run 1 | 10,7239 |
| Run 2 | 8,0483 |
| Run 3 | 9,9010 |
| Run 4 | 7,8895 |
| Run 5 | 10,3093 |
| Run 6 | 15,2091 |
| Run 7 | 10,1781 |
| Run 8 | 10,1266 |
| Run 9 | 12,4224 |
| Run 10 | 9,5923 |

---

## Resultados: SIN OCSP (Baseline)

| Métrica | Valor |
|---------|-------|
| **Latencia promedio** | 9,4528 ms/handshake |
| **Desviación estándar** | 2,0775 ms |
| **IC 95% (inferior)** | 7,9667 ms |
| **IC 95% (superior)** | 10,9389 ms |
| **Rango de confianza** | [7,9667 - 10,9389] ms |

### Detalle de las 10 corridas

| Corrida | Latencia (ms) |
|---------|---------------|
| Run 1 | 8,2645 |
| Run 2 | 10,4987 |
| Run 3 | 6,4103 |
| Run 4 | 10,9890 |
| Run 5 | 6,3492 |
| Run 6 | 12,0846 |
| Run 7 | 8,2474 |
| Run 8 | 10,8401 |
| Run 9 | 9,2166 |
| Run 10 | 11,6279 |

---

## Análisis de Overhead OCSP

| Métrica | Valor |
|---------|-------|
| **Diferencia absoluta** | 0,9872 ms |
| **Diferencia porcentual** | 10,44 % |

### Interpretación

⚠️ **ADVERTENCIA:** OCSP añade aproximadamente **0,9872ms** por handshake

Esto representa un incremento del **10,44%** en la latencia.

**Recomendaciones:**
- Considerar implementar OCSP stapling para reducir el impacto
- Evaluar si la validación de revocación es crítica para tu caso de uso
- Implementar caching de respuestas OCSP

---

## Comparativa Visual

### Latencia Promedio

```
Con OCSP:    10,4400 ms
Sin OCSP:    9,4528 ms
Overhead:    0,9872 ms (10,44%)
```

### Rango de Confianza (95%)

**Con OCSP:**    [8,9288 - 11,9513] ms
**Sin OCSP:**    [7,9667 - 10,9389] ms

---

## Conclusiones

- Se realizaron **10 corridas** de **3 segundos** cada una por escenario
- Se calcularon intervalos de confianza del 95% usando distribución t-Student
- El análisis incluye desviación estándar y error estándar

