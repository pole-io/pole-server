#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
image="${POLE_IMAGE:-pole-control-plane:local}"
docker_arch="$(docker info --format '{{.Architecture}}')"

case "${docker_arch}" in
  aarch64|arm64) goarch=arm64 ;;
  x86_64|amd64) goarch=amd64 ;;
  *)
    printf 'unsupported Docker architecture: %s\n' "${docker_arch}" >&2
    exit 1
    ;;
esac

printf '[pole-k8s] build console web\n'
(
  cd "${root_dir}/console/web"
  npm ci --legacy-peer-deps
  npm run build
)

printf '[pole-k8s] build linux/%s binary\n' "${goarch}"
(
  cd "${root_dir}"
  CGO_ENABLED=0 GOOS=linux GOARCH="${goarch}" go build -tags nomsgpack -o pole-server \
    -ldflags="-X github.com/pole-io/pole-server/pkg/common/version.Version=$(cat version 2>/dev/null || printf dev) -X github.com/pole-io/pole-server/pkg/common/version.BuildDate=$(date '+%Y%m%d.%H%M%S')" .
)

printf '[pole-k8s] build image %s\n' "${image}"
docker build -f "${root_dir}/deploy/kubernetes/Dockerfile" -t "${image}" "${root_dir}"
