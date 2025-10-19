#!/bin/bash

# 格式化 go.mod
go mod tidy -compat=1.17

# docker run -t --rm -v $(pwd):/app -w /app golangci/golangci-lint:v2.0.2 golangci-lint run -v --timeout 30m

# 处理 go imports 的格式化
if [ ! -d "style_tool" ]; then
    mkdir -p style_tool
    cd style_tool

    is_arm=$(/usr/bin/uname -m | grep -E "arm|aarch64" | wc -l)
    goimports_target_file="goimports-reviser_3.9.1_linux_amd64.tar.gz"

    if [ "$(uname)" == "Darwin" ]; then
        if [ "${is_arm}" == "1" ]; then
            goimports_target_file="goimports-reviser_3.9.1_darwin_arm64.tar.gz"
        else
            goimports_target_file="goimports-reviser_3.9.1_darwin_amd64.tar.gz"
        fi
    fi

    wget "https://github.com/incu6us/goimports-reviser/releases/download/v3.9.1/${goimports_target_file}"
    tar -zxvf ${goimports_target_file}
    mv goimports-reviser ../
    cd ../
fi

# 处理 go 代码格式化
go fmt ./...

find . -name "*.go" -type f | grep -v .pb.go | grep -v test/tools/tools.go | grep -v ./plugin.go |
    xargs -I {} ./goimports-reviser -rm-unused -company-prefixes github.com/pole-io/specification -project-name github.com/pole-io/pole-server -format {}
