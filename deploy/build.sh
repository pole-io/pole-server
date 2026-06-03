#!/bin/bash
# Tencent is pleased to support the open source community by making Polaris available.
#
# Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
#
# Licensed under the BSD 3-Clause License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# https://opensource.org/licenses/BSD-3-Clause
#
# Unless required by applicable law or agreed to in writing, software distributed
# under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
# CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

set -e

if [[ $(uname) == 'Darwin' ]]; then
  realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
  }

  md5sum() {
    md5 $*
  }
fi

workdir=$(dirname $(dirname $(realpath $0)))
version=$(cat version 2>/dev/null)
bin_name="pole-server"

if [ "${GOOS}" == "windows" ]; then
  bin_name="pole-server.exe"
fi

if [ "${GOOS}" == "" ]; then
  GOOS=$(go env GOOS)
fi

if [ "${GOARCH}" == "" ]; then
  GOARCH=$(go env GOARCH)
fi

if [ $# == 1 ]; then
  version=$1
fi
if [ $# == 2 ]; then
  version=$1
  export GOARCH=$2
fi

folder_name="pole-server-release_${version}.${GOOS}.${GOARCH}"
pkg_name="${folder_name}.zip"
echo "GOOS is ${GOOS}, GOARCH is ${GOARCH}, binary name is ${bin_name}"

echo "workdir=${workdir}"
cd ${workdir}

# 清理环境
rm -rf ${folder_name}
rm -f "${pkg_name}"
rm -f ${bin_name}

# 禁止 CGO_ENABLED 参数打开
export CGO_ENABLED=0

build_date=$(date "+%Y%m%d.%H%M%S")
package="github.com/pole-io/pole-server/pkg/common/version"
sqldb_res="plugin/store/mysql"
console_web_dir="console/web"
GOARCH=${GOARCH} GOOS=${GOOS} go build -o ${bin_name} -ldflags="-X ${package}.Version=${version} -X ${package}.BuildDate=${build_date}"

if [ -d "${console_web_dir}" ]; then
  echo "build console web"
  (
    cd "${console_web_dir}"
    npm ci --legacy-peer-deps
    npm run build
  )
fi

# 打包
mkdir -p ${folder_name}
cp ${bin_name} ${folder_name}
mkdir -p ${folder_name}/${sqldb_res}
cp -r ${sqldb_res}/scripts/* ${folder_name}/${sqldb_res}
cp -r ./deploy/tools ${folder_name}/
cp -r ./deploy/conf ${folder_name}/
if [ -d "${console_web_dir}/dist" ]; then
  mkdir -p ${folder_name}/${console_web_dir}
  cp -r ${console_web_dir}/dist ${folder_name}/${console_web_dir}/
else
  echo "console web dist not found"
  exit 1
fi
zip -r "${pkg_name}" ${folder_name}
md5sum ${pkg_name} >"${pkg_name}.md5sum"
