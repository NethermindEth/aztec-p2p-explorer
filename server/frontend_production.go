//go:build production

package server

import (
	"net/http"

	"github.com/NethermindEth/aztec-p2p-explorer/frontend"
)

func Frontend() http.FileSystem {
	return frontend.FS()
}
