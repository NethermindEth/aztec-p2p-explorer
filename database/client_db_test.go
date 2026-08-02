package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Assumes that the test database defined in the docker compose file is running.
func TestNew(t *testing.T) {
	t.Run("Successful Connection", func(t *testing.T) {
		cfg := &Config{
			DatabaseSourceName: "postgres://explorer:explorer@localhost:15432/explorer?sslmode=disable",
		}

		db, err := New(cfg)
		require.NoError(t, err)
		require.NotNil(t, db)

		err = db.Ping()
		assert.NoError(t, err)

		defer db.Close()
	})

	t.Run("Invalid DSN Format", func(t *testing.T) {
		cfg := &Config{
			DatabaseSourceName: "invalid-dsn-format",
		}

		db, err := New(cfg)
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}
