# Reporte Benchmark OCSP

**Fecha:** 07/07/2026 11:54:21

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
| **Latencia promedio** | 10,1403 ms/handshake |
| **Desviación estándar** | 1,3215 ms |
| **IC 95% (inferior)** | 9,1950 ms |
| **IC 95% (superior)** | 11,0855 ms |
| **Rango de confianza** | [9,1950 - 11,0855] ms |

### Detalle de las 10 corridas

| Corrida | Latencia (ms) |
|---------|---------------|
| Run 1 | 8,5653 |
| Run 2 | 11,5607 |
| Run 3 | 10,7239 |
| Run 4 | 9,6154 |
| Run 5 | 10,9890 |
| Run 6 | 8,6580 |
| Run 7 | 8,1466 |
| Run 8 | 11,6279 |
| Run 9 | 11,3636 |
| Run 10 | 10,1523 |

---

## Resultados: SIN OCSP (Baseline)

| Métrica | Valor |
|---------|-------|
| **Latencia promedio** | 9,4588 ms/handshake |
| **Desviación estándar** | 2,3411 ms |
| **IC 95% (inferior)** | 7,7842 ms |
| **IC 95% (superior)** | 11,1334 ms |
| **Rango de confianza** | [7,7842 - 11,1334] ms |

### Detalle de las 10 corridas

| Corrida | Latencia (ms) |
|---------|---------------|
| Run 1 | 9,8280 |
| Run 2 | 9,8765 |
| Run 3 | 9,1743 |
| Run 4 | 8,0972 |
| Run 5 | 14,6520 |
| Run 6 | 8,6022 |
| Run 7 | 8,0000 |
| Run 8 | 8,0321 |
| Run 9 | 11,9048 |
| Run 10 | 6,4205 |

---

## Análisis de Overhead OCSP

| Métrica | Valor |
|---------|-------|
| **Diferencia absoluta** | 0,6815 ms |
| **Diferencia porcentual** | 7,20 % |

### Interpretación

⚠️ **ADVERTENCIA:** OCSP añade aproximadamente **0,6815ms** por handshake

Esto representa un incremento del **7,20%** en la latencia.

**Recomendaciones:**
- Considerar implementar OCSP stapling para reducir el impacto
- Evaluar si la validación de revocación es crítica para tu caso de uso
- Implementar caching de respuestas OCSP

---

## Comparativa Visual

### Latencia Promedio

```
Con OCSP:    10,1403 ms
Sin OCSP:    9,4588 ms
Overhead:    0,6815 ms (7,20%)
```

### Rango de Confianza (95%)

**Con OCSP:**    [9,1950 - 11,0855] ms
**Sin OCSP:**    [7,7842 - 11,1334] ms

---

## Conclusiones

- Se realizaron **10 corridas** de **3 segundos** cada una por escenario
- Se calcularon intervalos de confianza del 95% usando distribución t-Student
- El análisis incluye desviación estándar y error estándar

