#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
manifest_dir="${root_dir}/deploy/kubernetes"
namespace="pole-system"
image="${POLE_IMAGE:-pole-control-plane:local}"
shared_greptime_namespace="tidemind"
shared_greptime_deployment="maas-greptimedb-frontend"
shared_greptime_service="maas-greptimedb-frontend.tidemind.svc.cluster.local"
greptime_database="pole_observability"

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

current_greptime_type="$(kubectl -n "${namespace}" get service pole-greptimedb \
  -o jsonpath='{.spec.type}' 2>/dev/null || true)"
current_greptime_target="$(kubectl -n "${namespace}" get service pole-greptimedb \
  -o jsonpath='{.spec.externalName}' 2>/dev/null || true)"
if [[ -n "${current_greptime_type}" ]] && {
  [[ "${current_greptime_type}" != "ExternalName" ]] ||
    [[ "${current_greptime_target}" != "${shared_greptime_service}" ]]
}; then
  kubectl -n "${namespace}" delete service pole-greptimedb
fi

kubectl apply -f "${manifest_dir}/greptimedb.yaml"
kubectl -n "${shared_greptime_namespace}" rollout status \
  "deployment/${shared_greptime_deployment}" --timeout=180s

kubectl -n "${namespace}" delete pod pole-greptimedb-init \
  --ignore-not-found --wait=true
kubectl -n "${namespace}" run pole-greptimedb-init \
  --image=curlimages/curl:8.12.1 \
  --restart=Never \
  --attach \
  --rm \
  --command -- \
  sh -ec "curl -fsS --data-urlencode 'sql=CREATE DATABASE IF NOT EXISTS ${greptime_database}' http://pole-greptimedb:4000/v1/sql"

kubectl apply -f "${manifest_dir}/otel-collector.yaml"
sed "s|image: pole-control-plane:local|image: ${image}|" \
  "${manifest_dir}/pole-control-plane.yaml" | kubectl apply -f -
kubectl apply -f "${manifest_dir}/gateway-route.yaml"

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

kubectl -n "${namespace}" rollout status deployment/pole-otel-collector --timeout=180s
kubectl -n "${namespace}" rollout status deployment/pole-control-plane --timeout=240s

# The old standalone StatefulSet is removed only after both callers have
# switched successfully. Kubernetes retains its PVC for rollback or export.
kubectl -n "${namespace}" delete statefulset pole-greptimedb \
  --ignore-not-found --wait=true

kubectl -n tidemind wait \
  --for=jsonpath='{.status.parents[0].conditions[?(@.type=="Accepted")].status}'=True \
  httproute/pole-console --timeout=60s
kubectl -n tidemind wait \
  --for=jsonpath='{.status.parents[0].conditions[?(@.type=="ResolvedRefs")].status}'=True \
  httproute/pole-console --timeout=60s

kubectl -n "${namespace}" get pods,services,persistentvolumeclaims
kubectl -n tidemind get httproute pole-console
