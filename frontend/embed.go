//go:build production

package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var frontend embed.FS

func FS() http.FileSystem {
	subbed, err := fs.Sub(frontend, "dist")
	if err != nil {
		panic(err)
	}
	return http.FS(subbed)
}
