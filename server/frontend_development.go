//go:build development

package server

import (
	"fmt"
	"net/http"
)

func init() {
	fmt.Println("Development server started")
}

func Frontend() http.FileSystem {
	return http.Dir("./frontend/dist")
}
