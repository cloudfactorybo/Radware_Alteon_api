#!/bin/sh
set -e

DEST="traefik/certs"
mkdir -p "$DEST"

if [ -f "$DEST/cert.pem" ] && [ -f "$DEST/key.pem" ] && [ "$1" != "--force" ]; then
    echo "certificados ya existen en $DEST (usa --force para regenerar)"
    exit 0
fi

# Hostname del equipo (CN). Override con HOST=... si quieres uno específico.
SHORT_HOST="$(hostname 2>/dev/null || echo localhost)"
HOST="${HOST:-$SHORT_HOST}"

# Nombres DNS a incluir en el SAN.
DNS_NAMES="$HOST $SHORT_HOST localhost alteon-api"
FQDN="$(hostname -f 2>/dev/null || true)"
[ -n "$FQDN" ] && DNS_NAMES="$DNS_NAMES $FQDN"

# IPs configuradas del host (globales, sin loopback ni link-local).
IPS="127.0.0.1"
if command -v ip >/dev/null 2>&1; then
    IPS="$IPS $(ip -o -4 addr show scope global 2>/dev/null | awk '{print $4}' | cut -d/ -f1)"
    IPS="$IPS $(ip -o -6 addr show scope global 2>/dev/null | awk '{print $4}' | cut -d/ -f1)"
else
    IPS="$IPS $(hostname -I 2>/dev/null || true)"
fi

# Construye subjectAltName deduplicando entradas.
SAN=""
add_san() {
    case ",$SAN," in
        *",$1,"*) return ;;
    esac
    [ -z "$SAN" ] && SAN="$1" || SAN="$SAN,$1"
}

for n in $DNS_NAMES; do
    [ -n "$n" ] && add_san "DNS:$n"
done
for i in $IPS; do
    [ -n "$i" ] && add_san "IP:$i"
done

openssl req -x509 -nodes -newkey rsa:2048 \
    -keyout "$DEST/key.pem" \
    -out    "$DEST/cert.pem" \
    -subj   "/CN=$HOST" \
    -days   3650 \
    -addext "subjectAltName=$SAN"

chmod 600 "$DEST/key.pem"
echo "certificados generados en $DEST"
echo "  CN:  $HOST"
echo "  SAN: $SAN"
