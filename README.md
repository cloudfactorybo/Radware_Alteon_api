# alteon-api

API REST en Go para consultar uno o más load balancers Radware Alteon. Desplegada en Docker con Postgres, Redis y Traefik (HTTPS en el mismo puerto que usa la app).

---

## 1. Prerequisitos

En el host donde vas a correr esto:

- `docker` y `docker compose` (Docker 20.10+).
- `openssl` (para generar el cert TLS autofirmado).
- `curl` + `jq` (opcional, para probar).

Verifica:

```bash
docker --version
docker compose version
openssl version
```

---

## 2. Preparación (antes del `docker compose up`)

### 2.1. Copiar y editar `.env`

```bash
cp .env.example .env
vi .env
```

Campos a revisar:

| Variable | Qué es | Recomendación |
|---|---|---|
| `POSTGRES_PASSWORD` | Password de la DB interna | Cámbialo a algo fuerte |
| `REDIS_PASSWORD` | Password de Redis | Opcional, pero mejor ponerlo |
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` | `info` está bien |
| `ALLOWED_ORIGINS` | CORS — orígenes web permitidos | `*` si no tienes frontend web |
| `PUBLIC_HTTPS_PORT` | Puerto público (Traefik HTTPS) | `5687` por defecto |

### 2.2. Generar certificado TLS autofirmado

```bash
./scripts/gen-certs.sh
```

Esto crea `traefik/certs/cert.pem` y `traefik/certs/key.pem` válidos para `localhost`, `alteon-api` y `127.0.0.1`.

Si el host tiene un nombre DNS específico:

```bash
HOST=mi-server.local ./scripts/gen-certs.sh --force
```

### 2.3. (Opcional) Decidir si la API es sólo localhost o expuesta a la LAN

Por defecto el compose bindea Traefik sólo a `127.0.0.1` — sólo se puede acceder desde el propio host. Si quieres exponer a la red, edita `docker-compose.yml`:

```yaml
    ports:
      - "127.0.0.1:${PUBLIC_HTTPS_PORT:-5687}:5687"   # sólo localhost (default)
      # - "${PUBLIC_HTTPS_PORT:-5687}:5687"           # toda la LAN
```

---

## 3. Levantar Docker

### 3.1. Build + arranque

```bash
docker compose up -d --build
```

Esto construye la imagen del app (multi-stage Go) y arranca 4 contenedores:

- `postgres` (base de datos)
- `redis` (cache)
- `app` (tu API Go)
- `traefik` (HTTPS en el puerto público)

### 3.2. Verificar que todo esté arriba

```bash
docker compose ps
```

Deberías ver los 4 servicios con estado `Up` o `healthy`.

### 3.3. Ver logs en vivo

```bash
# sólo el app
docker compose logs -f app

# todo
docker compose logs -f

# traefik solo (si algo falla con TLS)
docker compose logs -f traefik
```

---

## 4. Configuración inicial (después del `docker compose up`)

La DB arranca vacía: no hay alteons ni tokens. Tienes que sembrarla con el CLI `alteon-admin`.

### 4.1. (Opcional) Crear alias para comodidad

```bash
alias admin='docker compose exec app alteon-admin'
```

### 4.2. Agregar los alteons

```bash
docker compose exec app alteon-admin add-alteon Yape2 https://172.31.163.18 api 'CloudFactory2025.'
docker compose exec app alteon-admin add-alteon Yap1  https://172.31.163.37 api 'CloudFactory2025.'
```

Verifica:

```bash
docker compose exec app alteon-admin list-alteons
```

### 4.3. Crear un Bearer token

```bash
docker compose exec app alteon-admin create-token mi-cliente
```

**Copia el token inmediatamente** — se muestra sólo esa vez (se guarda hasheado en DB, no se puede recuperar).

### 4.4. Reiniciar el app para que cargue los alteons recién agregados

El server refresca la lista de alteons cada 5 minutos automáticamente, pero si quieres verlos *ya*:

```bash
docker compose restart app
```

---

## 5. Probar la API

```bash
export TOKEN=<pega aquí el token del paso 4.3>

# Health simple (sin auth)
curl -k https://127.0.0.1:5687/health

# Health deep — pinguea cada alteon
curl -k https://127.0.0.1:5687/health/deep | jq

# Endpoints v1 (requieren Bearer)
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/system         | jq
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/licenses       | jq
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/virtualservers | jq
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/monitoring     | jq
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/servicemap     | jq
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/gateways       | jq
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/smartnat       | jq
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/wanlinkgroups  | jq
curl -k -H "Authorization: Bearer $TOKEN" https://127.0.0.1:5687/api/v1/wanlinks       | jq
```

El endpoint `/gateways` expone el balanceo de enlaces de cada alteon en dos partes:

- `gateways` — default gateways (`IpCurCfgGwTable`): IP del enlace, métrica de balanceo (`ipCurCfgGwMetric`), health-check (`interval`/`retry`), estado administrativo, ARP, VLAN y prioridad.
- `interfaces` — interfaces IP (`IpCurCfgIntfTable`): IP/máscara, VLAN, estado, `peer` (interfaz del par HA) y `description` (nombre del ISP, p.ej. AXS/TIGO). Las interfaces se correlacionan con los gateways por VLAN.

Se devuelve siempre el valor entero crudo (`state`, `arp`, `metric`) junto a su nombre best-effort, porque el REST del vADC no documenta el enum. **Nota:** es config administrativa, no el estado operativo up/down del enlace — eso no lo expone el REST de este firmware (requeriría SNMP traps o syslog).

### Endpoints de balanceo de enlaces (LLB)

El LLB de Alteon se modela con grupos SLB (WAN Link Groups), real servers (WAN Links) y reglas Smart NAT:

- `/smartnat` — `SlbCurCfgSmartNatTable` (mapeo `localIp`→`dnatIp`, `wanLink`) + `SlbStatLinkpfSmartNATTable` (sesiones vivas). Cada regla trae `id`, `currSessions`, `totalSessions`, `type`, `localIp`, `dnatIp`, `wanLink`, `dnatPersist`.
- `/wanlinkgroups` — `SlbStatEnhGroupTable` (live: `currSessions`, `totalSessions`, `highestSessions`, `hcOctets`) + metric/backup de `SlbCurCfgEnhGroupTable`. `currSessions` = "Concurrent Connections" del GUI. `hcOctets`/`totalMB` es el **contador acumulado desde el último boot/clear**, no la ventana reciente que muestra el GUI.
- `/wanlinks` — devuelve `perId` (`SlbStatLinkpfRServerTable`) y `perIp` (`SlbStatLinkpfIpTable`), las dos subvistas del GUI. Cada fila trae sesiones (`currSessions`), ancho de banda actual/pico/total en Mbps (`upBwCurr`/`dnBwCurr`/`totBwCurr`, `*Peak`, `*Tot`) — strings, con `"--"` en `*Usage` cuando no hay límite configurado. **`perId` incluye `state`/`stateName` operativo del enlace** (`SlbStatLinkpfRServerTable.State`, enum runtime: 1=Running, 2=Failed, 3=Disabled, 4=Blocked) — este es el up/down real del enlace WAN.

Las tablas `SlbStatLinkpf*` (LinkProof) son las que alimentan las pestañas Smart NAT / WAN Links de la WebUI; se obtuvieron de las URLs REST que llama el GUI.

`-k` es necesario porque el cert es autofirmado. En producción reemplaza `traefik/certs/*.pem` con un cert real y quita el `-k`.

Status code + tiempo sin body:

```bash
curl -k -o /dev/null -w 'HTTP %{http_code}  %{time_total}s\n' \
  -H "Authorization: Bearer $TOKEN" \
  https://127.0.0.1:5687/api/v1/system
```

---

## 6. Operaciones comunes

### Reiniciar

```bash
docker compose restart app          # sólo el app
docker compose restart               # todos los servicios
```

### Rebuild (después de cambiar código Go)

```bash
docker compose up -d --build app
```

### Parar todo (preserva datos)

```bash
docker compose down
```

### Parar y **borrar todos los datos** (cuidado)

```bash
docker compose down -v
```

Esto borra los volúmenes `pgdata` y `redisdata` — pierdes alteons y tokens.

### Ver uso de recursos

```bash
docker compose stats
```

---

## 7. Administración día a día

### Alteons

```bash
# Agregar
docker compose exec app alteon-admin add-alteon <name> <url> <user> <pass>

# Listar
docker compose exec app alteon-admin list-alteons

# Deshabilitar temporalmente (no se borra, sólo no se consulta)
docker compose exec app alteon-admin disable-alteon <name>
docker compose exec app alteon-admin enable-alteon  <name>

# Borrar
docker compose exec app alteon-admin remove-alteon <name>
```

Los cambios se reflejan en la API en máx 5 min (warmup ticker). Para forzar ya: `docker compose restart app`.

### Tokens

```bash
# Emitir nuevo
docker compose exec app alteon-admin create-token <nombre-cliente>

# Listar (sólo metadata: id, nombre, creado, último uso, revocado)
docker compose exec app alteon-admin list-tokens

# Revocar por id
docker compose exec app alteon-admin revoke-token <id>
```

Los tokens se guardan **hasheados** (SHA-256). El valor en claro sólo se muestra en `create-token`. Si se pierde, revócalo y emite uno nuevo.

---

## 8. Troubleshooting

### El app no arranca

```bash
docker compose logs app
```

Causas comunes:
- Postgres aún no está healthy → el app reintenta. Espera unos segundos.
- Redis inalcanzable → revisa el password en `.env`.
- `DATABASE_URL` mal formado → no lo toques a mano, lo arma compose con las vars de Postgres.

### `/health/deep` devuelve `total: 0` pero `list-alteons` sí los muestra

El server aún no hizo refresh después de que los agregaste. Reinicia:

```bash
docker compose restart app
```

### 401 Unauthorized en endpoints `/api/v1/*`

Falta o es inválido el header Bearer. Verifica:

```bash
docker compose exec app alteon-admin list-tokens
```

Si el token está revocado, emite uno nuevo. Si no aparece, fue a otra DB (¿borraste los volúmenes?).

### Error TLS al hacer curl

Si ves `SSL certificate problem: self-signed certificate`, agrega `-k` al curl. Para resolverlo en cliente, confía el `traefik/certs/cert.pem` en tu sistema o reemplaza con un cert emitido por una CA de confianza.

### Puerto 5687 ocupado

Cambia `PUBLIC_HTTPS_PORT` en `.env` y haz `docker compose up -d` de nuevo.

---

## 9. Estructura del proyecto

```
.
├── cmd/
│   ├── server/          # API HTTP (alteon-api)
│   └── admin/           # CLI de administración (alteon-admin)
├── internal/
│   ├── cache/           # Cliente Redis (cache con TTL)
│   ├── config/          # Config desde env vars
│   ├── handler/         # Handlers HTTP
│   ├── middleware/      # logging, gzip, cors, auth
│   ├── models/          # Structs JSON
│   ├── service/         # Lógica del alteon
│   └── storage/         # Repos Postgres (alteons + tokens)
├── pkg/httpclient/      # Cliente HTTP compartido
├── traefik/
│   ├── dynamic.yml      # TLS config
│   └── certs/           # cert.pem + key.pem
├── scripts/
│   └── gen-certs.sh     # Generador de cert autofirmado
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## 10. Variables de entorno

| Variable | Default | Descripción |
|---|---|---|
| `SERVER_HOST` | `127.0.0.1` | Bind interno del app (en compose: `0.0.0.0`) |
| `SERVER_PORT` | `5687` | Puerto interno del app (en compose: `8080`) |
| `DATABASE_URL` | `postgres://alteon:alteon@localhost:5432/alteon?sslmode=disable` | DSN de Postgres |
| `REDIS_ADDR` | `localhost:6379` | Host:port de Redis |
| `REDIS_PASSWORD` | *(vacío)* | Password de Redis |
| `REDIS_DB` | `0` | Número de DB Redis |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `ALLOWED_ORIGINS` | `*` | CORS, lista CSV (ej. `https://foo,https://bar`) |
| `AUTH_DISABLED` | *(vacío)* | `true` apaga el Bearer token (sólo dev) |
| `PUBLIC_HTTPS_PORT` | `5687` | Puerto público de Traefik (en `.env`, usado por compose) |
