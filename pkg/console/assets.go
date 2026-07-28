//go:build !consoleassets

package console

import (
	"embed"
	"io/fs"
)

// consoleAssets 供普通开发期 Go 工具使用。正式制品必须带 consoleassets
// 构建标签，并由 scripts/build-console-assets.sh 先生成完整前端产物。
//
//go:embed all:internal/assets/dist
var consoleAssets embed.FS

func embeddedConsoleAssets() fs.FS {
	assets, err := fs.Sub(consoleAssets, "internal/assets/dist")
	if err != nil {
		panic(err)
	}
	return assets
}
