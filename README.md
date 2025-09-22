# Admira ETL + API (Go)

Sistema de ETL y API para procesamiento de datos de marketing digital con métricas de conversión.
---
## Descripción

Este servicio ingiere datos de **Ads** y **CRM** (usando Mocky como fuente de datos), los normaliza, cruza por UTM y expone métricas de marketing digital a través de una API REST. 

Las características principales incluyen:

- **Idempotencia** en la ingesta de datos
- **Timeouts** y reintentos con backoff exponencial
- **Health checks** para monitoreo
- **Métricas calculadas**: CPC, CPA, CVRs, ROAS
- **Filtros avanzados** por fecha, canal, UTM

---

## Requisitos

- **Go**: 1.22 o superior
- **Docker**: Opcional (para contenedores)
- **Memoria**: Mínimo 512MB RAM

---

## Instalación

### Opción 1: Desde Código Fuente

```bash
# Clonar el repositorio
git clone <repository-url>
cd admira-go

# Instalar dependencias
go mod tidy

# Compilar
go build -o admira cmd/api/main.go
```

### Opción 2: Docker

```bash
# Construir imagen
docker build -t admira-go .

# O usar docker-compose
docker-compose up -d
```

---

## Configuración

### Variables de Entorno

Crea un archivo `.env` basado en `.env.example`:

```bash
# URLs de las APIs externas
ADS_API_URL=https://mocki.io/v1/9dcc2981-2bc8-465a-bce3-47767e1278e6
CRM_API_URL=https://mocki.io/v1/6a064f10-829d-432c-9f0d-24d5b8cb71c7

# Configuración del servidor
PORT=8080

# (Opcional) Secreto para export
SINK_SECRET=your-secret-key
```

### Valores por Defecto

| Variable | Valor por Defecto | Descripción |
|----------|-------------------|-------------|
| `ADS_API_URL` | Mocky URL | API de datos publicitarios |
| `CRM_API_URL` | Mocky URL | API de oportunidades CRM |
| `PORT` | `8080` | Puerto del servidor |
| `SINK_SECRET` | - | Secreto para export (opcional) |

---

## Ejecución

### Desarrollo Local

```bash
# Opción 1: Go directo
go run cmd/api/main.go

# Opción 2: Binario compilado
./admira

# Opción 3: Con variables de entorno
PORT=9090 go run cmd/api/main.go
```

### Docker

```bash
# Imagen única
docker run -p 8080:8080 admira-go

# Docker Compose
docker-compose up -d
```

### Verificar Instalación

```bash
# Health check
curl http://localhost:8080/healthz
# Respuesta: ok

# Readiness check
curl http://localhost:8080/readyz
# Respuesta: ready
```

---

## Endpoints

### Health Checks

| Endpoint | Método | Descripción | Respuesta |
|----------|--------|-------------|-----------|
| `/healthz` | GET | Liveness probe | `200 OK` |
| `/readyz` | GET | Readiness probe | `200 OK` |

### ETL

| Endpoint | Método | Descripción | Parámetros |
|----------|--------|-------------|------------|
| `/ingest/run` | POST | Ejecutar proceso ETL | `since` (opcional) |

### Métricas

| Endpoint | Método | Descripción | Parámetros |
|----------|--------|-------------|------------|
| `/metrics/channel` | GET | Métricas por canal | Ver tabla de filtros |

#### Filtros de Métricas

| Parámetro | Tipo | Descripción | Ejemplo |
|-----------|------|-------------|---------|
| `from` | string | Fecha inicio (YYYY-MM-DD) | `2024-01-01` |
| `to` | string | Fecha fin (YYYY-MM-DD) | `2024-01-31` |
| `channel` | string | Canal publicitario | `google` |
| `campaign_id` | string | ID de campaña | `summer2024` |
| `utm_campaign` | string | UTM Campaign | `summer_sale` |
| `utm_source` | string | UTM Source | `google` |
| `utm_medium` | string | UTM Medium | `cpc` |
| `limit` | int | Límite de resultados | `100` |
| `offset` | int | Offset para paginación | `0` |

---

## Ejemplos de Uso

### 1. Ejecutar ETL (Ingesta de Datos)

```bash
# Ingesta completa
curl -X POST http://localhost:8080/ingest/run

# Ingesta desde fecha específica
curl -X POST "http://localhost:8080/ingest/run?since=2024-01-01"
```

**Respuesta:**
```json
{
  "status": "ok"
}
```

### 2. Obtener Métricas por Canal

```bash
# Métricas básicas
curl "http://localhost:8080/metrics/channel"

# Con filtros de fecha
curl "http://localhost:8080/metrics/channel?from=2024-01-01&to=2024-01-31"

# Filtrado por canal específico
curl "http://localhost:8080/metrics/channel?channel=google&limit=10"

# Filtrado por UTM
curl "http://localhost:8080/metrics/channel?utm_campaign=summer_sale&utm_source=google"
```

**Respuesta:**
```json
{
  "count": 2,
  "items": [
    {
      "date": "2024-01-15",
      "channel": "google",
      "campaign_id": "summer2024",
      "clicks": 150,
      "impressions": 5000,
      "cost": 75.50,
      "leads": 12,
      "opportunities": 8,
      "closed_won": 3,
      "revenue": 1500.00,
      "cpc": 0.50,
      "cpa": 6.29,
      "cvr_lead_to_opp": 0.67,
      "cvr_opp_to_won": 0.38,
      "roas": 19.87
    }
  ]
}
```

### 3. Health Checks

```bash
# Liveness check
curl http://localhost:8080/healthz
# Respuesta: ok

# Readiness check  
curl http://localhost:8080/readyz
# Respuesta: ready
```

### Ejemplos con Postman

#### Collection JSON para Postman:

```json
{
  "info": {
    "name": "Admira API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/healthz",
          "host": ["{{base_url}}"],
          "path": ["healthz"]
        }
      }
    },
    {
      "name": "Run ETL",
      "request": {
        "method": "POST",
        "header": [],
        "url": {
          "raw": "{{base_url}}/ingest/run",
          "host": ["{{base_url}}"],
          "path": ["ingest", "run"]
        }
      }
    },
    {
      "name": "Get Channel Metrics",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/metrics/channel?from=2024-01-01&to=2024-01-31&limit=10",
          "host": ["{{base_url}}"],
          "path": ["metrics", "channel"],
          "query": [
            {"key": "from", "value": "2024-01-01"},
            {"key": "to", "value": "2024-01-31"},
            {"key": "limit", "value": "10"}
          ]
        }
      }
    }
  ],
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080"
    }
  ]
}
```

---

## Decisiones de Diseño

### 1. Arquitectura Simplicada
- **Decisión**: Usar solo stdlib de Go
- **Razón**: Reducir dependencias externas y complejidad
- **Trade-off**: Menos funcionalidades out-of-the-box

### 2. Almacenamiento en Memoria
- **Decisión**: Store en memoria para MVP
- **Razón**: Simplicidad y velocidad de desarrollo
- **Trade-off**: Sin persistencia entre reinicios

### 3. ETL Secuencial
- **Decisión**: Procesamiento secuencial de datos
- **Razón**: Simplicidad y evitar condiciones de carrera
- **Trade-off**: Menor throughput que procesamiento paralelo

### 4. Join por UTM con Fallback
- **Decisión**: Join primario por (utm_campaign, utm_source, utm_medium)
- **Razón**: Precisión en la atribución
- **Fallback**: Solo utm_campaign si faltan source/medium

### 5. Idempotencia por Claves Naturales
- **Decisión**: Upsert basado en claves naturales
- **Razón**: Permitir reprocesamiento seguro
- **Implementación**: Ads=(date,campaign_id,channel), CRM=(opportunity_id)

### 6. Timeouts y Retry
- **Decisión**: 3 intentos con backoff exponencial
- **Razón**: Resiliencia ante fallos temporales
- **Configuración**: 300ms, 600ms, 1200ms

---

## Limitaciones

### Limitaciones Actuales (MVP)

| Área | Limitación | Impacto | Solución Futura |
|------|------------|---------|-----------------|
| **Persistencia** | Solo memoria | Sin durabilidad | PostgreSQL/ClickHouse |
| **Escalabilidad** | Single instance | Limitado por RAM | Clustering/Kubernetes |
| **Concurrencia** | ETL secuencial | Throughput limitado | Worker pools |
| **UTM Fallback** | Solo campaign | Posible sobre-atribución | Mejorar lógica de join |
| **Monitoreo** | Logs básicos | Visibilidad limitada | Prometheus/Grafana |

### Limitaciones Técnicas

- **Memoria**: Limitada por RAM disponible
- **Cardinalidad**: No optimizado para millones de registros
- **Concurrencia**: ETL no paraleliza por campaña
- **Observabilidad**: Métricas Prometheus opcionales

### Limitaciones de Datos

- **UTM Incompletos**: Fallback puede causar sobre-atribución
- **Monedas**: Sin soporte para múltiples monedas
- **Timezones**: Solo UTC

---

## Desarrollo

### Estructura del Proyecto

```
admira-go/
├── cmd/api/           # Punto de entrada
├── internal/
│   ├── http/          # Servidor HTTP y rutas
│   ├── ingest/        # ETL y clientes HTTP
│   ├── model/         # Estructuras de datos
│   ├── store/         # Almacenamiento en memoria
│   ├── metrics/       # Cálculo de métricas
│   └── util/          # Utilidades
├── test/              # Tests
├── go.mod
├── Dockerfile
└── docker-compose.yml
```

---


<div align="center">

</div>
