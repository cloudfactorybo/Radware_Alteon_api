# Alteon API Gateway — Multi-Instance

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.25.3-00ADD8?style=for-the-badge&logo=go)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker)
![Postgres](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis)
![Traefik](https://img.shields.io/badge/Traefik-v3.6-24A1C1?style=for-the-badge&logo=traefikproxy)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Status](https://img.shields.io/badge/Status-Production-success?style=for-the-badge)

**API Gateway RESTful para gestión centralizada de múltiples instancias Radware Alteon**

[Características](#-características) •
[Quick Start](#-quick-start) •
[Endpoints](#-endpoints-api) •
[Arquitectura](#-arquitectura) •
[Admin CLI](#-admin-cli)

</div>

---

## 📋 Tabla de Contenido

- [Descripción General](#-descripción-general)
- [Características](#-características)
- [Novedades de esta versión](#-novedades-de-esta-versión)
- [Requisitos](#-requisitos)
- [Quick Start](#-quick-start)
- [Endpoints API](#-endpoints-api)
  - [Health Check](#1-health-check)
  - [Health Deep](#2-health-deep)
  - [System Information](#3-system-information)
  - [Licenses](#4-licenses)
  - [Virtual Servers](#5-virtual-servers)
  - [Monitoring](#6-monitoring)
  - [Service Map](#7-service-map)
- [Autenticación](#-autenticación)
- [Admin CLI](#-admin-cli)
- [Arquitectura](#-arquitectura)
  - [Diagrama de Componentes](#diagrama-de-componentes)
  - [Flujo de Peticiones](#flujo-de-peticiones)
  - [Esquema de Base de Datos](#esquema-de-base-de-datos)
  - [Estructura del Proyecto](#estructura-del-proyecto)
- [Configuración](#-configuración)
- [Middlewares](#-middlewares)
- [Cache (Redis)](#-cache-redis)
- [TLS / Certificados](#-tls--certificados)
- [Logging](#-logging)
- [Operaciones día a día](#-operaciones-día-a-día)
- [Desarrollo](#-desarrollo)
- [Troubleshooting](#-troubleshooting)
- [Licencia](#-licencia)

---

## 🎯 Descripción General

**Alteon API Gateway** es un servicio RESTful escrito en Go que expone una API unificada sobre múltiples instancias de **Radware Alteon** (Application Delivery Controllers). En esta versión la app ya **no es un único binario systemd con config hardcodeada**: corre dentro de un stack Docker con **Postgres + Redis + Traefik** y la lista de Alteons se administra en caliente desde una **CLI dedicada** (`alteon-admin`).

El servicio permite:
- 🔄 **Consultas concurrentes** a N Alteons en paralelo (goroutines + `sync.WaitGroup`)
- 🗄️ **Configuración dinámica** — Alteons en Postgres, refresco automático cada 5 minutos
- 🔐 **Autenticación Bearer token** (SHA-256 hashed en DB) con CLI para emitir/revocar
- ⚡ **Cache Redis** con TTL para reducir presión sobre los Alteons
- 🌐 **HTTPS terminado en Traefik** (cert autofirmado o reemplazable)
- 📝 **Logs estructurados** (JSON o formato humano) con `X-Request-ID` correlation
- 🩺 **Health endpoints** (simple + deep) para K8s / load balancers
- 🛡️ **Graceful shutdown** y timeouts agresivos para evitar leaks

---

## ✨ Características

| Característica | Descripción |
|---------------|-------------|
| **Multi-Instancia dinámica** | Alteons gestionados en Postgres, modificables sin reiniciar |
| **Stack Docker** | `postgres` · `redis` · `app` · `traefik` orquestados con Compose |
| **Bearer auth** | Tokens 32-byte hex, hash SHA-256 en DB, revocación + last_used tracking |
| **Cache Redis** | TTL 15s para llamadas a `/config/Slb*` (stats y real-server info) |
| **Concurrencia controlada** | Semáforo de 8 requests concurrentes por Alteon |
| **Auto-refresh** | Ticker cada 5 min refresca lista desde DB y precalienta service map |
| **HTTP/2 + GZIP** | Compresión content-type-aware (solo texto/JSON/XML/JS) |
| **Request IDs** | `X-Request-ID` propagado a todos los logs internos |
| **Logs duales** | Formato `text` legible (default) o `json` para ingesta |
| **CORS configurable** | `ALLOWED_ORIGINS` con allowlist o wildcard |
| **TLS terminado** | Traefik 3.x sirve HTTPS en `:5687`, certs en `traefik/certs/` |
| **Graceful shutdown** | 30s para drenar requests vivas en SIGINT/SIGTERM |
| **Tests** | Unit tests para parsers, formatters y cleanup logic |

---

## 🆕 Novedades de esta versión

> Cambios respecto a la versión anterior (binario systemd + config en código):

- ❌ **Eliminado**: instalación systemd (`install-alteon-api.sh`), config hardcoded en `internal/config/config.go`, single-binary deployment.
- ✅ **Agregado**:
  - 🐳 Stack **Docker Compose** completo (`docker-compose.yml`, `Dockerfile` multi-stage).
  - 🗄️ Persistencia en **Postgres 16** con migración automática al arranque.
  - ⚡ Cache **Redis 7** con cliente `redis/go-redis/v9`.
  - 🌐 Reverse proxy **Traefik v3.6** con TLS terminado.
  - 🔐 Sistema de **Bearer tokens** con hashing SHA-256 + revocación.
  - 🛠 Binario admin **`alteon-admin`** (`cmd/admin/`) con 8 comandos.
  - 📡 **`/health/deep`** que pinguea cada Alteon.
  - 🆕 Prefijo de versión **`/api/v1/*`** (con auth) — `/health` queda sin auth.
  - 🪪 Middleware **`AuthMiddleware`** validando contra DB.
  - 🔁 **Refresh ticker** que recarga Alteons desde DB cada 5 minutos.
  - 🆔 **`X-Request-ID`** generado por request, expuesto en headers y logs.
  - 🎨 **Custom text formatter** alineado por columnas (más legible que JSON en local).
  - 📦 Tipo **`FlexString`** para manejar campos del Alteon que oscilan entre `"123"` y `123`.
  - 🔄 **Retry logic** en `/api/v1/servicemap` (8 intentos con backoff lineal) — el statdb del Alteon a veces tarda en estar listo.
  - 🧹 **`cleanServiceMap`** elimina valores por defecto (`"OK"`, `"Not Available"`) para reducir payload.
  - 🌐 IPv4/IPv6 parsing en `extractIPFromURL` con `url.Hostname()`.
  - 📊 Wrapper de respuesta **`{data, errors}`** — agregación parcial con detalle de fallas por Alteon.
  - ⚙ **Server timeouts** explícitos (`ReadHeader=10s`, `Read=30s`, `Write=90s`, `Idle=120s`).
  - 🧪 Suite de tests (`internal/service/alteon_test.go`).

---

## 📦 Requisitos

| Componente | Versión |
|-----------|---------|
| **Docker** | 20.10+ |
| **Docker Compose** | v2 |
| **OpenSSL** | cualquiera (para generar el cert TLS) |
| **Red** | Conectividad HTTPS a las instancias Alteon |
| **Puerto** | `5687` libre (configurable vía `PUBLIC_HTTPS_PORT`) |

### Dependencias Go (compiladas en la imagen)

```go
require (
    github.com/gorilla/mux         v1.8.1   // Router HTTP
    github.com/jackc/pgx/v5        v5.9.2   // Driver Postgres
    github.com/redis/go-redis/v9   v9.18.0  // Cliente Redis
    github.com/sirupsen/logrus     v1.9.3   // Logging estructurado
)
```

---

## 🚀 Quick Start

### 1. Preparar entorno

```bash
# Clonar repo
git clone <repo-url>
cd Api-Alteon

# Configurar variables
cp .env.example .env
vi .env   # ajusta POSTGRES_PASSWORD, REDIS_PASSWORD, ALLOWED_ORIGINS, etc.

# Generar cert TLS autofirmado
./scripts/gen-certs.sh
# → crea traefik/certs/cert.pem y traefik/certs/key.pem (válidos 10 años)
```

### 2. Levantar el stack

```bash
docker compose up -d --build
docker compose ps
```

Deberías ver 4 contenedores (`postgres`, `redis`, `app`, `traefik`) en estado `Up` o `healthy`.

### 3. Sembrar Alteons + token

```bash
# Agregar Alteons (la DB arranca vacía)
docker compose exec app alteon-admin add-alteon ALTEON-01 https://192.168.42.110 api 'pass'
docker compose exec app alteon-admin add-alteon ALTEON-02 https://192.168.42.111 api 'pass'

# Emitir un token (cópialo YA, no se puede recuperar)
docker compose exec app alteon-admin create-token mi-cliente

# (opcional) forzar refresh inmediato — el ticker corre cada 5 min
docker compose restart app
```

### 4. Probar

```bash
export TOKEN=<token del paso 3>

curl -k https://127.0.0.1:5687/health
curl -k https://127.0.0.1:5687/health/deep | jq
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/system | jq
```

> `-k` es necesario porque el cert es autofirmado. En producción, reemplaza `traefik/certs/*.pem` por un cert emitido por una CA y quita el `-k`.

---

## 🌐 Endpoints API

Base URL: `https://<host>:5687`

| Método | Path | Auth | Descripción |
|--------|------|------|-------------|
| `GET` | `/health` | ❌ | Liveness probe |
| `GET` | `/health/deep` | ❌ | Pinguea cada Alteon (readiness) |
| `GET` | `/api/v1/system` | ✅ | Info de sistema agregada |
| `GET` | `/api/v1/licenses` | ✅ | Licencias + capacidad |
| `GET` | `/api/v1/virtualservers` | ✅ | VServers + estadísticas |
| `GET` | `/api/v1/monitoring` | ✅ | CPU, memoria, cores |
| `GET` | `/api/v1/servicemap` | ✅ | Mapa VServer → Group → RealServer |

> Todos los endpoints `/api/v1/*` requieren `Authorization: Bearer <token>` salvo que `AUTH_DISABLED=true` (solo dev).

### Estructura de respuesta agregada

Los endpoints `/api/v1/*` devuelven siempre el mismo wrapper:

```json
{
  "data": [ /* resultados por Alteon */ ],
  "errors": [
    { "alteon": "ALTEON-02", "error": "auth 401: ..." }
  ]
}
```

- `data`: solo Alteons que respondieron OK.
- `errors`: omitido si todos respondieron bien.
- Si **ninguno** respondió → HTTP `502 Bad Gateway` con `data: []` y errores poblados.

---

### 1. Health Check

Liveness probe simple. No toca DB ni Alteons.

```http
GET /health
```

```json
{ "status": "healthy" }
```

| Status | Significado |
|--------|-------------|
| `200` | Servicio arriba |

---

### 2. Health Deep

Lanza un `Ping` (`GET /config?prop=sysName`) en paralelo a cada Alteon registrado.

```http
GET /health/deep
```

```json
{
  "status": "healthy",
  "total": 2,
  "ok": 2,
  "alteons": [
    { "alteon": "ALTEON-01", "ok": true },
    { "alteon": "ALTEON-02", "ok": true }
  ]
}
```

| Estado global | Cuándo |
|---------------|--------|
| `healthy`     | todos los Alteons respondieron — `200` |
| `degraded`    | algunos fallaron — `200` (con detalle) |
| `unhealthy`   | ninguno respondió — `503` |

---

### 3. System Information

```http
GET /api/v1/system
Authorization: Bearer <token>
```

Devuelve por Alteon: `sysName`, uptime, RTC, memoria MP, IPs de mgmt (IPv4 + 4 SLAAC IPv6), MAC y estado FIPS.

```json
{
  "data": [
    {
      "alteonName": "ALTEON-01",
      "alteonUrl":  "https://192.168.42.110",
      "alteonIp":   "192.168.42.110",
      "sysName":    "ALTEON-01",
      "agSwitchUpTime": "11 days, 6:30:45",
      "mpMemStatsFree": 2048576,
      "mpMemStatsTotal": 4194304,
      "agMgmtCurCfgIpAddr": "192.168.42.110",
      "agMgmtCurCfgMask":   "255.255.255.0",
      "hwMACAddress":       "00:11:22:33:44:55"
    }
  ]
}
```

---

### 4. Licenses

```http
GET /api/v1/licenses
Authorization: Bearer <token>
```

Combina dos tablas del Alteon (`AgLicenseInfoTable` + `AgLicenseCapacityInfoTable`) y enriquece con:
- `expirationDate` parseado desde el string `"Expires on MM/DD/YY"`
- `daysUntilExpiration` calculado
- `capacitySizeFormatted` (`"Unlimited"`, `"500 Mbps"`, `"5 Gbps"`, `"Not Applicable"`)

```json
{
  "data": [{
    "alteonName": "ALTEON-01",
    "licenses": [{
      "licenseIdx": 1,
      "softwareKey": "SSL",
      "status": "Active. Expires on 12/31/27",
      "expirationDate": "12/31/27",
      "daysUntilExpiration": 605,
      "capacitySize": 5000,
      "capacitySizeFormatted": "5 Gbps",
      "currentUsage": "1200",
      "peakUsage": "3800",
      "hasCapacityInfo": true
    }]
  }]
}
```

---

### 5. Virtual Servers

```http
GET /api/v1/virtualservers
Authorization: Bearer <token>
```

VServers con servicios anidados y, por servicio, estadísticas del real server (cacheadas 15s en Redis).

```json
{
  "data": [{
    "alteonName": "ALTEON-01",
    "virtualServers": [{
      "index": "1",
      "currSessions": 124,
      "totalSessions": 9890,
      "highestSessions": 312,
      "services": [{
        "virtServIndex": "1",
        "svcIndex": 1,
        "vport": 443,
        "rport": 443,
        "state": 2,
        "stateName": "Running",
        "real_server": {
          "realStatus": 1,
          "realStatusName": "Running",
          "macAddr": "aa:bb:cc:dd:ee:ff",
          "ipAddr":  "10.0.0.5",
          "thruput": 12345,
          "totalBw": "987654",
          "pktPerSec": 230
        }
      }]
    }]
  }]
}
```

| `state` / `realStatus` | `stateName` |
|------------------------|-------------|
| `1` | `Blocked` (state) / `Running` (realStatus) |
| `2` | `Running` / `Failed` |
| `3` | `Failed` / `Disabled` |
| `4` | `Disabled` / `Blocked` |
| `5` | `Slowstart` |

> Campos como `totalBw` y `serverRtt` usan el tipo **`FlexString`** porque el Alteon a veces los devuelve como número y a veces como string — siempre se serializan como string en la salida.

---

### 6. Monitoring

```http
GET /api/v1/monitoring
Authorization: Bearer <token>
```

```json
{
  "data": [{
    "alteonName": "ALTEON-01",
    "cpu":    { "util1Second": 22, "util4Seconds": 18, "util64Seconds": 15 },
    "memory": {
      "totalMemory": 4194304,
      "initConfigMemory": 1048576,
      "usedMemory": 1048576,
      "availableMemory": 3145728,
      "usagePercentage": 25.0,
      "safetyMargin1": 524288,
      "safetyMargin2": 262144
    },
    "cores": [{
      "index": 0,
      "curProcSize": 102400,
      "memPressStat": 0,
      "memUseFrom1stMargin": 50,
      "peakUsageFrom1stMargin": 65
    }]
  }]
}
```

`usagePercentage` se calcula localmente como `(initConfigMemory / totalMemory) * 100`.

---

### 7. Service Map

```http
GET /api/v1/servicemap
Authorization: Bearer <token>
```

Topología completa: VServer → VService → RGroup → RealServers.

> ⚠️ El Alteon a veces devuelve `status: "err"` mientras el statdb se inicializa. El servicio reintenta hasta **8 veces** con backoff lineal antes de fallar.

```json
{
  "data": [{
    "alteonName": "ALTEON-01",
    "timestamp": 1736702394,
    "vservers": [{
      "id": "1",
      "ip": "10.0.0.100",
      "vservices": [{
        "name": "WEB-443",
        "vport": 443,
        "protocol": "https",
        "application": "ssl",
        "rgroup": {
          "id": "WEB-POOL",
          "rservers": [
            { "id": "rs1", "ip": "192.168.1.10", "rports": [8080] },
            { "id": "rs2", "ip": "192.168.1.11", "rports": [8080] }
          ]
        }
      }]
    }]
  }]
}
```

> Optimización: `cleanServiceMap()` elimina campos con valores por defecto (`cstatus: "OK"`, `hc_reason: "Not Available"`, `status: "ok"`) para reducir el tamaño del payload.

---

## 🔐 Autenticación

Todos los endpoints `/api/v1/*` esperan el header:

```
Authorization: Bearer <token>
```

### Cómo funciona

1. **Generación**: `alteon-admin create-token <nombre>` produce 32 bytes random → hex (64 chars) → se imprime UNA vez.
2. **Storage**: Se guarda **`SHA-256(token)`** en la columna `api_tokens.token_hash`. El plain nunca toca la DB.
3. **Validación**: Cada request hashea el token recibido y busca match en DB con `revoked = false`.
4. **Tracking**: `last_used_at` se actualiza en cada validación exitosa.
5. **Revocación**: `alteon-admin revoke-token <id>` marca el token; el siguiente request falla con 401.

### Errores de auth

```json
401 Unauthorized
WWW-Authenticate: Bearer realm="alteon-api"

{ "error": "token inválido o revocado" }
```

Mensajes posibles: `"falta header Authorization"`, `"esperaba 'Bearer <token>'"`, `"token vacío"`, `"token inválido o revocado"`.

### Bypass (solo dev)

```bash
AUTH_DISABLED=true docker compose up -d
```

El servicio loguea `auth deshabilitado (AUTH_DISABLED=true)` al arrancar.

---

## 🛠 Admin CLI

`alteon-admin` es un binario incluido en la misma imagen Docker. Se usa con `docker compose exec app alteon-admin <cmd>`.

### Alteons

```bash
# Crear (insecure_skip_verify=true por default)
alteon-admin add-alteon <name> <url> <user> <pass>

# Listar
alteon-admin list-alteons

# Habilitar / deshabilitar (no se borra, solo se omite en consultas)
alteon-admin enable-alteon  <name>
alteon-admin disable-alteon <name>

# Eliminar
alteon-admin remove-alteon  <name>
```

Cambios visibles en máx **5 minutos** (warmup ticker). Para forzar inmediato: `docker compose restart app`.

### Tokens

```bash
# Emitir nuevo (imprime el plain UNA vez)
alteon-admin create-token <nombre-cliente>

# Listar (sólo metadata: id, nombre, creado, último uso, revocado)
alteon-admin list-tokens

# Revocar por id
alteon-admin revoke-token <id>
```

---

## 🏗 Arquitectura

### Diagrama de Componentes

```mermaid
graph TB
    subgraph Cliente
        A[HTTP Client]
    end

    subgraph "Stack Docker (red interna)"
        T[Traefik v3.6<br/>HTTPS :5687]
        APP[App Go<br/>:8080 interno]
        PG[(Postgres 16<br/>alteons + tokens)]
        RD[(Redis 7<br/>cache TTL 15s)]
    end

    subgraph "Backends Alteon"
        AL1[Alteon 01<br/>HTTPS]
        AL2[Alteon 02<br/>HTTPS]
        ALN[Alteon N...]
    end

    A -->|Bearer + HTTPS| T
    T -->|HTTP interno| APP

    APP <-->|migraciones + CRUD| PG
    APP <-->|GET/SET cache| RD

    APP -->|goroutine 1| AL1
    APP -->|goroutine 2| AL2
    APP -->|goroutine N| ALN

    style T fill:#24A1C1,color:#fff,stroke:#1A7B95
    style APP fill:#00ADD8,color:#fff,stroke:#005F87
    style PG fill:#336791,color:#fff,stroke:#1F3D54
    style RD fill:#DC382D,color:#fff,stroke:#8C231C
    style AL1 fill:#9B59B6,color:#fff,stroke:#6C3483
    style AL2 fill:#9B59B6,color:#fff,stroke:#6C3483
    style ALN fill:#9B59B6,color:#fff,stroke:#6C3483
```

### Flujo de Peticiones

```mermaid
sequenceDiagram
    participant Client
    participant Traefik
    participant App
    participant DB as Postgres
    participant Cache as Redis
    participant A1 as Alteon-01
    participant A2 as Alteon-02

    Client->>Traefik: HTTPS GET /api/v1/system + Bearer
    Traefik->>App: HTTP GET /api/v1/system
    App->>App: LoggingMiddleware (genera X-Request-ID)
    App->>App: GzipMiddleware
    App->>App: CORSMiddleware
    App->>DB: AuthMiddleware (validar token hasheado)
    DB-->>App: token válido — UPDATE last_used_at

    par Llamadas concurrentes
        App->>A1: HTTPS /config?prop=sysName,...
        App->>A2: HTTPS /config?prop=sysName,...
    end

    A1-->>App: JSON
    A2-->>App: JSON

    App->>App: Agregar resultados → {data, errors}
    App-->>Traefik: 200 + gzip
    Traefik-->>Client: 200 HTTPS
```

### Flujo de Cache (endpoints de stats)

```mermaid
sequenceDiagram
    participant App
    participant Cache as Redis
    participant Alteon

    App->>Cache: GET alteon:endpoint
    alt cache hit
        Cache-->>App: bytes
    else cache miss
        Cache-->>App: nil
        App->>Alteon: HTTPS GET /config/SlbEnhStat
        Alteon-->>App: JSON
        App->>Cache: SET key con TTL 15s
    end
```

### Esquema de Base de Datos

```mermaid
erDiagram
    ALTEONS {
        bigserial id PK
        text name UK
        text base_url
        text username
        text password
        bool insecure_skip_verify
        text ca_cert
        bool enabled
        timestamptz created_at
        timestamptz updated_at
    }
    API_TOKENS {
        bigserial id PK
        text token_hash UK
        text name
        timestamptz created_at
        timestamptz last_used_at
        bool revoked
    }
```

> Las migraciones corren idempotentemente al arrancar el server (`storage.Open`) — `CREATE TABLE IF NOT EXISTS`.

### Estructura del Proyecto

```
.
├── cmd/
│   ├── server/main.go        # Entry point del API HTTP (alteon-api)
│   └── admin/main.go         # CLI de administración (alteon-admin)
│
├── internal/
│   ├── cache/                # Cliente Redis (cache.Cache)
│   ├── config/               # Config desde env vars
│   ├── handler/              # Handlers HTTP (health, system, license, ...)
│   ├── logformat/            # Custom text formatter para logrus
│   ├── middleware/           # auth, logger, gzip, cors
│   ├── models/               # DTOs JSON + tipo FlexString
│   ├── reqctx/               # Request ID en context
│   ├── service/              # Lógica multi-alteon + retry/clean
│   └── storage/              # Repos Postgres (alteons, tokens)
│
├── pkg/
│   └── httpclient/           # Cliente HTTP compartido (TLS skip-verify)
│
├── traefik/
│   ├── dynamic.yml           # TLS config (minTLS 1.2)
│   └── certs/                # cert.pem + key.pem
│
├── scripts/
│   └── gen-certs.sh          # Generador de cert autofirmado
│
├── Dockerfile                # Multi-stage Go 1.25 → alpine
├── docker-compose.yml        # postgres + redis + app + traefik
└── .env.example
```

---

## ⚙ Configuración

### Variables de entorno (`.env`)

| Variable | Default | Descripción |
|----------|---------|-------------|
| `SERVER_HOST` | `127.0.0.1` (compose: `0.0.0.0`) | Bind interno del app |
| `SERVER_PORT` | `5687` (compose: `8080`) | Puerto interno del app |
| `DATABASE_URL` | `postgres://alteon:alteon@localhost:5432/alteon?sslmode=disable` | DSN Postgres |
| `REDIS_ADDR` | `localhost:6379` | Host:port Redis |
| `REDIS_PASSWORD` | *(vacío)* | Password Redis |
| `REDIS_DB` | `0` | DB Redis |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `text` | `text` (humano) o `json` (ingesta) |
| `ALLOWED_ORIGINS` | `*` | CORS, lista CSV (`https://foo,https://bar`) |
| `AUTH_DISABLED` | *(vacío)* | `true` desactiva el Bearer (sólo dev) |
| `PUBLIC_HTTPS_PORT` | `5687` | Puerto público de Traefik (en `.env` para compose) |
| `POSTGRES_USER` | `alteon` | Usuario DB |
| `POSTGRES_PASSWORD` | `alteon` | **Cámbialo a algo fuerte** |
| `POSTGRES_DB` | `alteon` | Nombre de la DB |

### Constantes en código

| Dónde | Constante | Valor | Sentido |
|-------|-----------|-------|---------|
| `cmd/server/main.go` | `warmupInitialDelay` | `5s` | Delay antes del primer warmup |
| `cmd/server/main.go` | `warmupInterval` | `5m` | Refresca lista + precalienta service map |
| `cmd/server/main.go` | `warmupTimeout` | `60s` | Timeout para 1 ciclo de warmup |
| `cmd/server/main.go` | `refreshTimeout` | `10s` | Timeout para query a DB de Alteons |
| `internal/service/alteon.go` | `maxConcurrentRequests` | `8` | Semáforo por Alteon |
| `internal/service/alteon.go` | `statsTTL` | `15s` | TTL de cache para stats endpoints |
| `internal/service/alteon.go` | `maxRetries` (servicemap) | `8` | Reintentos cuando statdb no listo |
| `pkg/httpclient/client.go` | `Timeout` | `30s` | Timeout HTTP por request al Alteon |
| `pkg/httpclient/client.go` | `MaxIdleConns` | `100` | Pool de conexiones HTTP |
| `pkg/httpclient/client.go` | `IdleConnTimeout` | `90s` | Vida de conexiones idle |
| `cmd/server/main.go` | `srv.ReadHeaderTimeout` | `10s` | Hardening contra slowloris |
| `cmd/server/main.go` | `srv.ReadTimeout` | `30s` | Read total |
| `cmd/server/main.go` | `srv.WriteTimeout` | `90s` | Write total (tolera service map lento) |
| `cmd/server/main.go` | `srv.IdleTimeout` | `120s` | Keepalive |

---

## 🔧 Middlewares

### Orden de ejecución (raíz → handler)

```
LoggingMiddleware → GzipMiddleware → CORSMiddleware → [AuthMiddleware solo en /api/v1] → Handler
```

### 1. Logging Middleware (`logger.go`)

- Genera un **`X-Request-ID`** (12 hex chars) por request y lo escribe en el header de respuesta.
- Inyecta el ID en `context.Context` para correlacionar logs internos del servicio.
- Captura status code + bytes via `statusRecorder`.
- **Niveles inteligentes**:
  - `5xx` → `error`
  - `4xx` → `warn`
  - `/health` 2xx → `debug` (es ruido de healthcheck cada pocos segundos)
  - resto → `info`
- Cliente IP: prioriza `X-Forwarded-For` (Traefik), luego `X-Real-IP`, luego `RemoteAddr`.

### 2. Gzip Middleware (`gzip.go`)

- Activa solo si el cliente envía `Accept-Encoding: gzip`.
- **Content-type-aware**: solo comprime `text/*`, `application/json`, `application/xml`, `application/javascript`. Imágenes, PDFs, etc. pasan sin tocar.
- Agrega `Vary: Accept-Encoding` al response.

### 3. CORS Middleware (`cors.go`)

- Si `ALLOWED_ORIGINS=*` → wildcard global.
- Si es lista CSV → match exacto contra el header `Origin` del request, agrega `Vary: Origin`.
- Maneja preflight `OPTIONS` con `204 No Content`.

### 4. Auth Middleware (`auth.go`)

- Solo en `/api/v1/*`.
- Valida `Authorization: Bearer <token>` contra hashes en `api_tokens`.
- Pone el `tokenID` en context (clave `middleware.CtxTokenID`) para auditoría futura.
- En 401 envía `WWW-Authenticate: Bearer realm="alteon-api"` (RFC 6750).

---

## ⚡ Cache (Redis)

Solo se cachean los endpoints **internos** del Alteon que el servicio llama por servicio dentro de un VServer (potencialmente decenas por petición):

| Endpoint Alteon | TTL | Key pattern |
|-----------------|-----|-------------|
| `/config/SlbEnhVirtServicesInfoTable/<idx>/` | 15s | `<alteonName>:<endpoint>` |
| `/config/SlbEnhStatVirtServiceTable/<v>/<s>/<r>?...` | 15s | `<alteonName>:<endpoint>` |
| `/config/SlbEnhRealServerInfoTable/<idx>?...` | 15s | `<alteonName>:<endpoint>` |

- **Política fail-open**: si Redis cae, el servicio sigue funcionando — solo pierde el hit rate (logs a nivel `debug`, no `error`).
- Pool: 20 conexiones, dial 3s, R/W 2s.
- Persistencia: AOF habilitado (`--appendonly yes`) en `redisdata` volume.

---

## 🔒 TLS / Certificados

```bash
# Generar nuevo cert (válido 10 años, RSA 2048)
./scripts/gen-certs.sh

# Forzar regeneración
HOST=mi-server.local ./scripts/gen-certs.sh --force
```

El cert generado tiene como SAN: `DNS:<HOST>`, `DNS:alteon-api`, `IP:127.0.0.1`.

Para reemplazar con un cert real (Let's Encrypt, CA corporativa):

```bash
cp /path/to/fullchain.pem traefik/certs/cert.pem
cp /path/to/privkey.pem   traefik/certs/key.pem
docker compose restart traefik
```

`traefik/dynamic.yml` fuerza **`minVersion: VersionTLS12`**.

---

## 📝 Logging

### Formato `text` (default)

```
18:48:07  INFO   GET  /api/v1/monitoring          200  14ms      req_id=f25eeafb client=172.18.0.1 bytes=591
18:48:07  DEBUG  GET  /health                     200  0.42ms    req_id=8a1c0e22 client=127.0.0.1 bytes=21
18:48:29  INFO   warmup service map                                ok=2 errors=0 duration=20.02s
18:50:05  ERROR  alteon call failed                                alteon=ALTEON-02 endpoint=system "status 406"
```

### Formato `json` (`LOG_FORMAT=json`)

```json
{"level":"info","method":"GET","path":"/api/v1/system","status":200,"duration_ms":14.21,"req_id":"f25eeafb","client":"172.18.0.1","bytes":591,"time":"2026-05-06T18:48:07Z","msg":"http request"}
```

### Comandos útiles

```bash
docker compose logs -f app                        # solo el app
docker compose logs -f                            # todo
docker compose logs -f traefik                    # solo Traefik (útil si TLS falla)
docker compose logs --since 5m app | grep ERROR
```

---

## 🔄 Operaciones día a día

### Reiniciar

```bash
docker compose restart app
docker compose restart                             # todos
```

### Rebuild tras cambio de código Go

```bash
docker compose up -d --build app
```

### Parar (preservando datos)

```bash
docker compose down
```

### Parar y borrar TODO (incluye DB y cache)

```bash
docker compose down -v
```

> ⚠️ Esto destruye `pgdata` y `redisdata` — se pierden Alteons y tokens. Vuelve a sembrarlos con `alteon-admin`.

### Forzar refresh de la lista de Alteons

```bash
docker compose restart app
```

(o esperar al ticker de 5 min).

### Ver consumo

```bash
docker compose stats
```

---

## 👨‍💻 Desarrollo

### Local sin Docker (necesitas Postgres + Redis externos)

```bash
go mod download

export DATABASE_URL='postgres://alteon:alteon@localhost:5432/alteon?sslmode=disable'
export REDIS_ADDR=localhost:6379
export AUTH_DISABLED=true
export LOG_LEVEL=debug
export LOG_FORMAT=text

# Server
go run ./cmd/server

# Admin (en otra terminal)
go run ./cmd/admin add-alteon LAB https://10.0.0.1 api 'pass'
go run ./cmd/admin create-token dev
```

### Tests

```bash
go test ./...                                  # todos
go test -race ./...                            # con race detector
go test -cover ./...                           # cobertura
go test -coverprofile=cov.out ./... && go tool cover -html=cov.out
```

Tests existentes (`internal/service/alteon_test.go`):
- `TestFormatCapacitySize` — formato `Mbps/Gbps/Unlimited`
- `TestParseExpirationDate_Valid/Invalid` — parsing de strings tipo `"Expires on 10/11/30"`
- `TestGetStateName` / `TestGetRealStatusName` — mappings numéricos
- `TestCleanServiceMap` / `TestCleanServiceMap_PreservesNonOK` — limpieza de defaults
- `TestExtractIPFromURL` — IPv4 / IPv6 / hostnames

### Build optimizado

El `Dockerfile` ya hace lo correcto, pero para builds locales:

```bash
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o alteon-api  ./cmd/server
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o alteon-admin ./cmd/admin
```

---

## 🔍 Troubleshooting

### `/health/deep` devuelve `total: 0` pero `list-alteons` los muestra

El warmup ticker corre cada 5 minutos. Para forzar:

```bash
docker compose restart app
```

### `401 Unauthorized` en `/api/v1/*`

```bash
# Verificar que el token existe y no está revocado
docker compose exec app alteon-admin list-tokens
```

Si no aparece, ¿se borró el volumen `pgdata`? Emite uno nuevo.

### Error TLS al hacer `curl`

`SSL certificate problem: self-signed certificate` → agrega `-k` al curl, o reemplaza `traefik/certs/*.pem` con un cert de CA confiable.

### Puerto 5687 ocupado

```bash
echo "PUBLIC_HTTPS_PORT=8443" >> .env
docker compose up -d traefik
```

### El app no arranca

```bash
docker compose logs app
```

Causas comunes:
- Postgres aún no `healthy` → el app reintenta solo, espera unos segundos.
- Redis inalcanzable → revisa `REDIS_PASSWORD` en `.env`.
- `DATABASE_URL` mal formado → no lo edites a mano, lo arma compose.

### `service map: statdb no listo después de 8 intentos`

El Alteon devuelve `status: "err"` cuando su statdb no terminó de inicializar tras un boot. El servicio reintenta 8 veces con backoff lineal (2s, 4s, 6s...). Si persiste:
1. `curl -k -u user:pass https://<alteon>/monitor/servicemap` directo desde el host.
2. Si falla ahí también, esperar / reiniciar el Alteon.

### Cache Redis inconsistente

```bash
docker compose exec redis redis-cli FLUSHDB
```

Solo afecta hit rate; no hay riesgo de corrupción (los datos de verdad están en el Alteon).

### `auth deshabilitado` aparece en logs en producción

Tienes `AUTH_DISABLED=true` en tu `.env` o variable de entorno. Quítalo:

```bash
unset AUTH_DISABLED
sed -i '/AUTH_DISABLED/d' .env
docker compose up -d app
```

---

## 📄 Licencia

Este proyecto está bajo la **Licencia MIT**. Ver archivo `LICENSE` para más detalles.

---

<div align="center">

**Desarrollado con ❤️ usando Go 1.25.3 · PostgreSQL 16 · Redis 7 · Traefik v3.6**

[⬆ Volver arriba](#alteon-api-gateway--multi-instance)

</div>
