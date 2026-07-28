#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
web_dir="${root_dir}/web/console"
target_dir="${root_dir}/pkg/console/internal/assets/dist"
mode="${1:-release}"

case "${mode}" in
  release) npm_command=build ;;
  test) npm_command=build:test ;;
  *)
    printf 'unsupported console asset mode: %s\n' "${mode}" >&2
    exit 1
    ;;
esac

if [[ "${POLE_CONSOLE_NPM_CI:-0}" == "1" || ! -d "${web_dir}/node_modules" ]]; then
  (
    cd "${web_dir}"
    npm ci --legacy-peer-deps
  )
fi

(
  cd "${web_dir}"
  npm run "${npm_command}"
)

mkdir -p "${target_dir}"
find "${target_dir}" -mindepth 1 ! -name placeholder.txt -delete
cp -R "${web_dir}/dist/." "${target_dir}/"

if [[ ! -f "${target_dir}/index.html" ]]; then
  printf 'console asset build did not produce index.html\n' >&2
  exit 1
fi
