### 🏗️ Componentes Principales

| Componente | Responsabilidad | Tecnología |
|------------|----------------|------------|
| **HTTP Client** | Comunicación con APIs externas | Go `net/http` |
| **ETL Process** | Transformación de datos | Go structs + JSON |
| **Memory Store** | Almacenamiento temporal | Go maps + RWMutex |
| **Metrics Calculator** | Cálculo de KPIs | Go algorithms |
| **REST API** | Exposición de datos | Go `net/http` |

---

## 2. Idempotencia y Reprocesamiento

### 🔑 Claves Naturales

| Entidad | Clave Natural | Propósito |
|---------|---------------|-----------|
| **Ads** | `(date, campaign_id, channel)` | Evitar duplicados por campaña |
| **CRM** | `(opportunity_id)` | Identificador único de oportunidad |

### ⚡ Operación de Upsert

```go
// Pseudocódigo de upsert
if exists(key) {
    update(record)
} else {
    insert(record)
}
```

### 🔄 Reprocesamiento
- ✅ **Seguro**: Reprocesar `/ingest/run` no duplica métricas
- ✅ **Eficiente**: Parámetro `since` evita trabajo innecesario
- ✅ **Idempotente**: Múltiples ejecuciones producen el mismo resultado

---

## 3. Particionamiento y Retención

### 📅 Estrategia de Particionamiento

#### **Desarrollo (Actual)**
- **Partición lógica**: Por fecha y canal
- **Almacenamiento**: Memoria (RAM)

#### **Producción (Futuro)**
- **Hot Data**: 90 días en DB/Cache
- **Histórico**: Data lake (S3/BigQuery)
- **Particiones**: `date=YYYY-MM-DD` y/o `channel`

### 💾 Retención de Datos

| Tipo | Retención | Almacenamiento |
|------|-----------|----------------|
| **Hot Data** | 90 días | PostgreSQL/ClickHouse |
| **Histórico** | 2+ años | S3/BigQuery |
| **Logs** | 30 días | CloudWatch/ELK |

---

## 4. Concurrencia y Throughput

### ⚡ Implementación Actual

| Componente | Concurrencia | Limitación |
|------------|--------------|------------|
| **HTTP Handlers** | ✅ Concurrente (Go) | Go runtime |
| **Memory Store** | ✅ RWMutex | Thread-safe |
| **ETL Process** | ❌ Secuencial | Simplicidad |

### �� Optimizaciones Futuras

```go
// Worker pool para ETL
type ETLWorkerPool struct {
    workers    int
    campaigns  chan Campaign
    results    chan Result
}
```

- **Worker Pools**: Fan-out por campaña/canal
- **Límite de Paralelismo**: Configurable por ambiente
- **Rate Limiting**: Control de carga en APIs externas

---

## 5. Calidad de Datos

### �� Normalización

| Campo | Transformación | Ejemplo |
|-------|----------------|---------|
| **Strings** | `lower()` + `trim()` | `"Google " → "google"` |
| **Fechas** | Parse seguro | `"2024-01-01" → time.Time` |
| **Números** | Negativos → 0 | `-5 → 0` |

### �� Estrategia de Join

#### **Join Primario**
```sql
-- Pseudocódigo SQL
JOIN ON (utm_campaign, utm_source, utm_medium)
```

#### **Fallback Strategy**
```go
if utm_source == "" || utm_medium == "" {
    // Fallback por utm_campaign solamente
    log.Warn("UTM incomplete, using campaign-only join")
}
```

### ��️ Protección de Datos
- **Divisiones**: Protegidas contra NaN/Inf
- **Registros inválidos**: Se omiten con log de advertencia
- **Validación**: Campos requeridos antes del procesamiento

---

## 6. Cálculo de Métricas

### 📊 Tabla de Métricas

| Métrica | Fórmula | Descripción | Unidad |
|---------|---------|-------------|--------|
| **CPC** | `cost / clicks` | Costo por clic | $ |
| **CPA** | `cost / leads` | Costo por lead | $ |
| **CVR Lead→Opp** | `opportunities / leads` | Conversión lead a oportunidad | % |
| **CVR Opp→Won** | `closed_won / opportunities` | Conversión oportunidad a venta | % |
| **ROAS** | `revenue / cost` | Retorno de inversión publicitaria | ratio |

### 🎯 Agregación

#### **Dimensiones**
- **Fecha**: `date`
- **Canal**: `channel`
- **Campaña**: `campaign_id`

#### **Filtros Disponibles**
- **Temporales**: `from`, `to`
- **Campaña**: `campaign_id`
- **UTM**: `utm_campaign`, `utm_source`, `utm_medium`
- **Paginación**: `limit`, `offset`

### 💰 Revenue Calculation
```go
// Revenue = Σ amount donde stage == "closed_won"
revenue := 0.0
for _, opp := range opportunities {
    if opp.Stage == "closed_won" {
        revenue += opp.Amount
    }
}
```

---

## 7. Observabilidad

### 🏥 Health Checks

| Endpoint | Propósito | Respuesta |
|----------|-----------|-----------|
| `/healthz` | Liveness probe | `200 OK` |
| `/readyz` | Readiness probe | `200 OK` |

### 📝 Logging

#### **Formato JSON Estructurado**
```json
{
  "level": "info",
  "request_id": "abc123",
  "endpoint": "/metrics/channel",
  "status": "ok",
  "duration_ms": 45,
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### **Campos Obligatorios**
- `level`: info, warn, error
- `request_id`: Identificador único
- `endpoint`: Ruta solicitada
- `status`: ok, error
- `duration_ms`: Tiempo de respuesta

### 📈 Métricas (Opcional)

#### **Prometheus**
- **Contadores**: Requests por endpoint
- **Histogramas**: Latencia de respuesta
- **Gauges**: Estado del ETL

#### **ETL Metrics**
- **Errores**: Contador de fallos
- **Reintentos**: Tasa de reintentos
- **Throughput**: Registros procesados/segundo

---

## 8. Resiliencia

### ⏱️ Timeouts y Retry

#### **Configuración de Timeouts**
```go
// HTTP Client
client := &http.Client{
    Timeout: 5 * time.Second,
}

// Context timeout
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()
```

#### **Estrategia de Reintentos**
```go
backoff := []time.Duration{
    300 * time.Millisecond,  // Intento 1
    600 * time.Millisecond,  // Intento 2
    1200 * time.Millisecond, // Intento 3
}
```

### 🔄 Política de Reintentos

| Código | Reintentar | Razón |
|--------|------------|-------|
| **2xx** | ❌ No | Éxito |
| **4xx** | ❌ No | Error del cliente |
| **5xx** | ✅ Sí | Error del servidor |
| **Timeout** | ✅ Sí | Problema de red |

### 🛡️ Circuit Breaker (Futuro)
```go
type CircuitBreaker struct {
    failureThreshold int
    timeout         time.Duration
    state          State // Closed, Open, HalfOpen
}
```

---

## 9. Seguridad

### �� Gestión de Secretos

#### **Variables de Entorno**
```bash
# .env
ADS_API_URL=https://api.ads.com
CRM_API_URL=https://api.crm.com
SINK_SECRET=your-secret-key
```

#### **Export con Firma**
```go
// HMAC-SHA256 signature
signature := hmac.New(sha256.New, secret)
signature.Write(payload)
xSignature := hex.EncodeToString(signature.Sum(nil))
```

### 🛡️ Seguridad en Producción

| Capa | Implementación | Propósito |
|------|----------------|-----------|
| **TLS** | Reverse proxy | Cifrado en tránsito |
| **Rate Limiting** | Nginx/CloudFlare | Protección DDoS |
| **Authentication** | JWT/OAuth2 | Autenticación |
| **Authorization** | RBAC | Control de acceso |

---

## 10. Evolución del Ecosistema

### ��️ Persistencia

#### **Migración de Store**
```go
// Interfaz de repositorio
type Repository interface {
    UpsertAds(ads AdsPerf) error
    UpsertCRM(crm CRMOpp) error
    GetAds(filters Filters) ([]AdsPerf, error)
    GetCRM(filters Filters) ([]CRMOpp, error)
}

// Implementaciones
type MemoryRepository struct { /* actual */ }
type PostgresRepository struct { /* futuro */ }
type ClickHouseRepository struct { /* futuro */ }
```

### �� Orquestación

#### **ETL Orchestration**
- **Airflow**: Workflows complejos
- **Argo Workflows**: Kubernetes-native
- **Kafka**: Streaming en tiempo real

#### **Contratos de Esquema**
```json
{
  "schema_version": "1.0",
  "ads_performance": {
    "required": ["date", "campaign_id", "channel"],
    "optional": ["utm_campaign", "utm_source", "utm_medium"]
  }
}
```

### �� API Evolution

#### **Versionado**
- **v1**: API actual
- **v2**: Nuevas funcionalidades
- **Deprecation**: Ciclo de vida controlado

#### **Testing**
- **Contract Tests**: Pact/OpenAPI
- **Load Tests**: K6/Gatling
- **Integration Tests**: Testcontainers

---

## 11. Fallos y Recuperación

### 🚨 Manejo de Errores

#### **Caídas de Fuentes**
```go
// ETL responde con error específico
if err := fetchData(); err != nil {
    return fmt.Errorf("ADS API unavailable: %w", err)
}
```

#### **Datos Corruptos**
```go
// Omitir filas problemáticas
for _, record := range records {
    if !isValid(record) {
        log.Warn("Skipping invalid record", "record", record)
        continue
    }
    process(record)
}
```

### �� Estrategias de Recuperación

| Escenario | Estrategia | Tiempo de Recuperación |
|-----------|------------|------------------------|
| **API Down** | Reintentos + fallback | 5-15 minutos |
| **Datos corruptos** | Skip + log | Inmediato |
| **Memoria llena** | Restart + re-ingest | 1-5 minutos |
| **Proceso crash** | Auto-restart | 30 segundos |

---

## 12. Limitaciones (MVP)

### ⚠️ Limitaciones Actuales

| Área | Limitación | Impacto |
|------|------------|---------|
| **Almacenamiento** | Solo memoria | Sin durabilidad |
| **Escalabilidad** | Single instance | Limitado por RAM |
| **UTM Fallback** | Solo campaign | Posible sobre-atribución |
| **Paralelización** | ETL secuencial | Throughput limitado |

### 🚀 Roadmap de Mejoras

#### **Fase 1** (Q1 2024)
- [ ] Persistencia en PostgreSQL
- [ ] Métricas Prometheus
- [ ] Tests unitarios completos

#### **Fase 2** (Q2 2024)
- [ ] ETL paralelizado
- [ ] Caché Redis
- [ ] API v2

#### **Fase 3** (Q3 2024)
- [ ] Data lake integration
- [ ] Real-time streaming
- [ ] ML-powered insights

---

## 13. Operación

### �� Despliegue

#### **Desarrollo Local**
```bash
# Opción 1: Go directo
go run cmd/api/main.go

# Opción 2: Docker Compose
docker-compose up -d
```

#### **Producción**
```bash
# Kubernetes
kubectl apply -f k8s/

# Docker Swarm
docker stack deploy -c docker-compose.prod.yml admira
```

### �� Monitoreo

#### **Métricas Clave**
- **Latencia P95**: < 200ms
- **Error Rate**: < 1%
- **ETL Success Rate**: > 99%
- **Memory Usage**: < 80%

#### **Alertas**
- **Error Rate** > 5% por 5 minutos
- **Latencia P95** > 1s por 2 minutos
- **ETL Failure** > 3 intentos consecutivos

### 🔧 Mantenimiento

#### **Logs**
- **Rotación**: Diaria
- **Retención**: 30 días
- **Formato**: JSON estructurado

#### **Métricas**
- **Retención**: 90 días
- **Granularidad**: 1 minuto
- **Storage**: Prometheus + Grafana

---

## 📚 Referencias

### 🔗 Enlaces Útiles
- [Go HTTP Server Best Practices](https://golang.org/doc/effective_go.html#web)
- [ETL Patterns](https://www.oreilly.com/library/view/etl-patterns/9781492043176/)
- [Observability in Go](https://github.com/golang/go/wiki/CodeReviewComments#logging)

### 📖 Documentación Adicional
- [API Documentation](./api-docs.md)
- [Deployment Guide](./deployment.md)
- [Troubleshooting](./troubleshooting.md)

---

<div align="center">

**🏗️ Admira ETL + API**  
*Sistema de procesamiento de datos de marketing digital*

[![Go Version](https://img.shields.io/badge/Go-1.22-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-MVP-orange.svg)]()

</div>