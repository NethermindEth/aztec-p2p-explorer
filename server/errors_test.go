package server_test

import (
	"fmt"
	"testing"

	"github.com/test-go/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/server"
)

func TestAPIError(t *testing.T) {
	e := server.NewAPIError(404, fmt.Errorf("not found"))

	require.Equal(t, 404, e.Status)
	require.Equal(t, "api error: 404 (not found) error=not found", e.Error())
}
