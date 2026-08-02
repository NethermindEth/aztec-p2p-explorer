package testutil

import (
	"testing"
)

func TestPrepareDatabase(t *testing.T) {
	db := prepareTestPostgresDatabase(t)
	db.Close()
}
