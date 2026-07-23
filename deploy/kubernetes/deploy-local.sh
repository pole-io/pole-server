#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
manifest_dir="${root_dir}/deploy/kubernetes"
namespace="pole-system"
image="${POLE_IMAGE:-pole-control-plane:local}"

mysql_user="${MYSQL_USER:-root}"
mysql_pwd="${MYSQL_PWD:?MYSQL_PWD must be set}"
console_jwt_secret="${POLE_CONSOLE_JWT_SECRET:?POLE_CONSOLE_JWT_SECRET must be set}"
auth_salt="${POLE_AUTH_SALT:?POLE_AUTH_SALT must be set}"

kubectl apply -f "${manifest_dir}/namespace.yaml"

system_secret_master_key="${POLE_SYSTEM_SECRET_MASTER_KEY:-}"
if [[ -z "${system_secret_master_key}" ]]; then
  system_secret_master_key="$(kubectl -n "${namespace}" get secret pole-runtime-secrets \
    -o jsonpath='{.data.SYSTEM_SECRET_MASTER_KEY}' 2>/dev/null | base64 --decode || true)"
fi
if [[ -z "${system_secret_master_key}" ]]; then
  system_secret_master_key="$(openssl rand -base64 32)"
fi

kubectl -n "${namespace}" create secret generic pole-runtime-secrets \
  --from-literal=MYSQL_USER="${mysql_user}" \
  --from-literal=MYSQL_PWD="${mysql_pwd}" \
  --from-literal=CONSOLE_JWT_SECRET="${console_jwt_secret}" \
  --from-literal=AUTH_SALT="${auth_salt}" \
  --from-literal=SYSTEM_SECRET_MASTER_KEY="${system_secret_master_key}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n "${namespace}" create configmap pole-runtime-config \
  --from-file=pole-server.yaml="${root_dir}/deploy/conf/pole-server.yaml" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -f "${manifest_dir}/mysql-external-service.yaml"
kubectl apply -f "${manifest_dir}/greptimedb.yaml"
kubectl apply -f "${manifest_dir}/otel-collector.yaml"
kubectl apply -f "${manifest_dir}/pole-control-plane.yaml"
kubectl apply -f "${manifest_dir}/gateway-route.yaml"

kubectl -n "${namespace}" set image deployment/pole-control-plane \
  pole-control-plane="${image}"

config_revision="$(shasum -a 256 "${root_dir}/deploy/conf/pole-server.yaml" | awk '{print $1}')"
image_revision="$(docker image inspect --format '{{.Id}}' "${image}" | shasum -a 256 | awk '{print $1}')"
secret_revision="$(printf '%s\0%s\0%s\0%s\0%s' "${mysql_user}" "${mysql_pwd}" "${console_jwt_secret}" "${auth_salt}" "${system_secret_master_key}" | shasum -a 256 | awk '{print $1}')"
kubectl -n "${namespace}" set env deployment/pole-control-plane \
  POLE_CONFIG_REVISION="${config_revision}" \
  POLE_IMAGE_REVISION="${image_revision}" \
  POLE_SECRET_REVISION="${secret_revision}"

collector_revision="$(shasum -a 256 "${manifest_dir}/otel-collector.yaml" | awk '{print $1}')"
kubectl -n "${namespace}" set env deployment/pole-otel-collector \
  POLE_CONFIG_REVISION="${collector_revision}"

kubectl -n "${namespace}" rollout status statefulset/pole-greptimedb --timeout=180s
kubectl -n "${namespace}" rollout status deployment/pole-otel-collector --timeout=180s
kubectl -n "${namespace}" rollout status deployment/pole-control-plane --timeout=240s

kubectl -n tidemind wait \
  --for=jsonpath='{.status.parents[0].conditions[?(@.type=="Accepted")].status}'=True \
  httproute/pole-console --timeout=60s
kubectl -n tidemind wait \
  --for=jsonpath='{.status.parents[0].conditions[?(@.type=="ResolvedRefs")].status}'=True \
  httproute/pole-console --timeout=60s

kubectl -n "${namespace}" get pods,services,persistentvolumeclaims
kubectl -n tidemind get httproute pole-console
