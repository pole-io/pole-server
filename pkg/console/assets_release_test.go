//go:build consoleassets

package console

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedConsoleAssetsContainIndexAndReferencedFiles(t *testing.T) {
	assets := embeddedConsoleAssets()
	index, err := fs.ReadFile(assets, "index.html")
	require.NoError(t, err)
	require.NotEmpty(t, index)

	references := regexp.MustCompile(`(?:src|href)="(/[^"]+)"`).FindAllSubmatch(index, -1)
	require.NotEmpty(t, references)
	for _, reference := range references {
		path := strings.TrimPrefix(string(reference[1]), "/")
		if path == "" {
			continue
		}
		_, err := fs.Stat(assets, path)
		require.NoErrorf(t, err, "index.html 引用的静态资源未被嵌入：%s", path)
	}
}
