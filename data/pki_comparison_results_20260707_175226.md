# Reporte Benchmark OCSP

**Fecha:** 07/07/2026 17:52:26

## Configuración del Benchmark

- **Herramienta:** OpenSSL s_time
- **Corridas por escenario:** 10
- **Duración por corrida:** 5 segundos
- **Protocolo:** TLS 1.3
- **Autenticación:** mTLS con PKI de un nivel

---

## Resultados: PKI de un nivel SIN OCSP (Baseline)

| Métrica | Valor |
|---------|-------|
| **Latencia promedio** | 5,6683 ms/handshake |
| **Desviación estándar** | 0,6384 ms |
| **IC 95% (inferior)** | 5,2117 ms |
| **IC 95% (superior)** | 6,1249 ms |
| **Rango de confianza** | [5,2117 - 6,1249] ms |

### Detalle de las 10 corridas

| Corrida | Latencia (ms) |
|---------|---------------|
| Run 1 | 6,2992 |
| Run 2 | 5,0891 |
| Run 3 | 6,2794 |
| Run 4 | 6,3694 |
| Run 5 | 4,8426 |
| Run 6 | 5,8737 |
| Run 7 | 6,1162 |
| Run 8 | 5,9347 |
| Run 9 | 5,0125 |
| Run 10 | 4,8662 |
---

## Comparativa Visual

### Latencia Promedio

```
PKI de un nivel sin OCSP:    5,6683 ms

### Rango de Confianza (95%)

**PKI de un nivel sin OCSP:**    [5,2117 - 6,1249] ms

---

## Conclusiones

- Se realizaron **10 corridas** de **5 segundos**
- Se calcularon intervalos de confianza del 95% usando distribución t-Student
- El análisis incluye desviación estándar y error estándar

