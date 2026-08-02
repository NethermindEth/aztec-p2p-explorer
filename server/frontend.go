//go:build !production && !development

// Default implementation used when no build tag is specified
package server

import (
	"net/http"
)

// Frontend returns a filesystem for serving static frontend assets
func Frontend() http.FileSystem {
	return http.Dir("./frontend/dist")
}
