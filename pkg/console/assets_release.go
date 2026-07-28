//go:build consoleassets

package console

import (
	"embed"
	"io/fs"
)

// consoleAssets 是 release/test Go 制品内嵌的完整 Console 前端。
// 显式列出 index.html，使未执行前端构建的制品编译在编译期失败。
//
//go:embed internal/assets/dist/index.html all:internal/assets/dist
var consoleAssets embed.FS

func embeddedConsoleAssets() fs.FS {
	assets, err := fs.Sub(consoleAssets, "internal/assets/dist")
	if err != nil {
		panic(err)
	}
	return assets
}
