#!/bin/sh

set -eu

src="${POLE_CONFIG_SOURCE:-/config-src/pole-server.yaml}"
dst="${POLE_CONFIG_RENDERED:-/work/pole-server.yaml}"

: "${MYSQL_USER:?MYSQL_USER must be set}"
: "${MYSQL_PWD:?MYSQL_PWD must be set}"
: "${MYSQL_HOST:?MYSQL_HOST must be set}"
: "${POLE_CONSOLE_JWT_SECRET:?POLE_CONSOLE_JWT_SECRET must be set}"
: "${POLE_AUTH_SALT:?POLE_AUTH_SALT must be set}"
: "${POD_IP:?POD_IP must be set}"
: "${OTEL_COLLECTOR_GRPC_ENDPOINT:?OTEL_COLLECTOR_GRPC_ENDPOINT must be set}"
: "${OTEL_COLLECTOR_HTTP_ENDPOINT:?OTEL_COLLECTOR_HTTP_ENDPOINT must be set}"
: "${GREPTIMEDB_HTTP_ENDPOINT:?GREPTIMEDB_HTTP_ENDPOINT must be set}"
: "${POLE_CLUSTER_NAME:?POLE_CLUSTER_NAME must be set}"

case "${MYSQL_MAX_OPEN_CONNS:-}" in ''|*[!0-9]*) echo 'MYSQL_MAX_OPEN_CONNS must be numeric' >&2; exit 1 ;; esac
case "${MYSQL_MAX_IDLE_CONNS:-}" in ''|*[!0-9]*) echo 'MYSQL_MAX_IDLE_CONNS must be numeric' >&2; exit 1 ;; esac

validate_value() {
  name="$1"
  value="$2"
  case "${value}" in
    *"
"*|*"$(printf '\r')"*)
      echo "${name} must be single-line" >&2
      exit 1
      ;;
  esac
  if printf '%s' "${value}" | LC_ALL=C grep -q '[[:cntrl:]]'; then
    echo "${name} must not contain control characters" >&2
    exit 1
  fi
}

validate_value MYSQL_USER "${MYSQL_USER}"
validate_value MYSQL_PWD "${MYSQL_PWD}"
validate_value MYSQL_HOST "${MYSQL_HOST}"
validate_value POLE_CONSOLE_JWT_SECRET "${POLE_CONSOLE_JWT_SECRET}"
validate_value POLE_AUTH_SALT "${POLE_AUTH_SALT}"
validate_value POD_IP "${POD_IP}"
validate_value OTEL_COLLECTOR_GRPC_ENDPOINT "${OTEL_COLLECTOR_GRPC_ENDPOINT}"
validate_value OTEL_COLLECTOR_HTTP_ENDPOINT "${OTEL_COLLECTOR_HTTP_ENDPOINT}"
validate_value GREPTIMEDB_HTTP_ENDPOINT "${GREPTIMEDB_HTTP_ENDPOINT}"
validate_value POLE_CLUSTER_NAME "${POLE_CLUSTER_NAME}"

yaml_quote() {
  value="$1"
  escaped="$(printf '%s' "${value}" | sed 's/\\/\\\\/g; s/"/\\"/g')"
  printf '"%s"' "${escaped}"
}

export MYSQL_USER_YAML="$(yaml_quote "${MYSQL_USER}")"
export MYSQL_PWD_YAML="$(yaml_quote "${MYSQL_PWD}")"
export MYSQL_HOST_YAML="$(yaml_quote "${MYSQL_HOST}")"
export MYSQL_DSN_YAML="$(yaml_quote "${MYSQL_USER}:${MYSQL_PWD}@tcp(${MYSQL_HOST})/pole_server?loc=Local")"
export POLE_CONSOLE_JWT_SECRET_YAML="$(yaml_quote "${POLE_CONSOLE_JWT_SECRET}")"
export POLE_AUTH_SALT_YAML="$(yaml_quote "${POLE_AUTH_SALT}")"
export POD_IP_YAML="$(yaml_quote "${POD_IP}")"
export OTEL_COLLECTOR_GRPC_ENDPOINT_YAML="$(yaml_quote "${OTEL_COLLECTOR_GRPC_ENDPOINT}")"
export OTEL_COLLECTOR_HTTP_ENDPOINT_YAML="$(yaml_quote "${OTEL_COLLECTOR_HTTP_ENDPOINT}")"
export GREPTIMEDB_HTTP_ENDPOINT_YAML="$(yaml_quote "${GREPTIMEDB_HTTP_ENDPOINT}")"
export POLE_CLUSTER_NAME_YAML="$(yaml_quote "${POLE_CLUSTER_NAME}")"

mkdir -p "$(dirname "${dst}")" /app/logs/runtime /app/data/observability/otel-events

sed \
  -e 's|dbUser: ##DB_USER##|dbUser: ${MYSQL_USER_YAML}|g' \
  -e 's|dbPwd: ##DB_PWD##|dbPwd: ${MYSQL_PWD_YAML}|g' \
  -e 's|dbAddr: ##DB_ADDR##|dbAddr: ${MYSQL_HOST_YAML}|g' \
  -e 's|dns: "${MYSQL_USER}:${MYSQL_PWD}@tcp(${MYSQL_HOST})/pole_server?loc=Local" ##DB_DNS##|dns: ${MYSQL_DSN_YAML}|g' \
  -e 's|secretKey: polarismesh@2021|secretKey: ${POLE_CONSOLE_JWT_SECRET_YAML}|g' \
  -e 's|salt: polarismesh@2021|salt: ${POLE_AUTH_SALT_YAML}|g' \
  -e 's|# self_address: 127.0.0.1|self_address: ${POD_IP_YAML}|g' \
  -e 's|endpoint: http://127.0.0.1:4000|endpoint: ${GREPTIMEDB_HTTP_ENDPOINT_YAML}|g' \
  -e 's|endpoint: 127.0.0.1:4317|endpoint: ${OTEL_COLLECTOR_GRPC_ENDPOINT_YAML}|g' \
  -e 's|logsEndpoint: http://127.0.0.1:4318/v1/logs|logsEndpoint: ${OTEL_COLLECTOR_HTTP_ENDPOINT_YAML}|g' \
  -e 's|maxOpenConns: 300|maxOpenConns: ${MYSQL_MAX_OPEN_CONNS}|g' \
  -e 's|maxIdleConns: 50|maxIdleConns: ${MYSQL_MAX_IDLE_CONNS}|g' \
  -e 's|cluster: local-docker|cluster: ${POLE_CLUSTER_NAME_YAML}|g' \
  "${src}" >"${dst}"

# The source file is maintained outside this image. Fail closed when its shape
# changes instead of silently starting with a default credential or endpoint.
for required in \
  '${MYSQL_USER_YAML}' \
  '${MYSQL_PWD_YAML}' \
  '${MYSQL_HOST_YAML}' \
  '${MYSQL_DSN_YAML}' \
  '${POLE_CONSOLE_JWT_SECRET_YAML}' \
  '${POLE_AUTH_SALT_YAML}' \
  '${POD_IP_YAML}' \
  '${GREPTIMEDB_HTTP_ENDPOINT_YAML}' \
  '${OTEL_COLLECTOR_GRPC_ENDPOINT_YAML}' \
  '${OTEL_COLLECTOR_HTTP_ENDPOINT_YAML}' \
  '${POLE_CLUSTER_NAME_YAML}'
do
  if ! grep -Fq "${required}" "${dst}"; then
    echo "rendered configuration is missing required placeholder: ${required}" >&2
    exit 1
  fi
done

if grep -Eq 'db(User|Pwd|Addr): ##DB_|##DB_DNS##|polarismesh@2021|# self_address: 127\.0\.0\.1' "${dst}"; then
  echo 'rendered configuration still contains an unsafe source default' >&2
  exit 1
fi

exec /app/pole-server "$@"
