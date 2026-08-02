package testutil

import (
	"database/sql"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"

	"github.com/NethermindEth/aztec-p2p-explorer/database"
)

func PrepareTestDatabase(t *testing.T) (*sql.DB, func()) {
	db := prepareTestPostgresDatabase(t)
	teardown := func() {
		db.Close()
	}
	return db, teardown
}

func prepareTestPostgresDatabase(t *testing.T) *sql.DB {
	// This config makes assumptions about using a tmpfs backed postgres running
	// from docker compose.
	conf := pgtestdb.Config{
		DriverName: "postgres",
		User:       "explorer",
		Password:   "explorer",
		Host:       "localhost",
		Port:       "15432",
		Options:    "sslmode=disable",
		TestRole: &pgtestdb.Role{
			Username:     pgtestdb.DefaultRole().Username,
			Password:     pgtestdb.DefaultRole().Password,
			Capabilities: "SUPERUSER",
		},
	}

	migrator := golangmigrator.New("migrations", golangmigrator.WithFS(database.Migrations))
	prepped := pgtestdb.Custom(t, conf, migrator)

	db, err := database.New(&database.Config{DatabaseSourceName: prepped.URL()})
	if err != nil {
		t.Fatal(err)
	}

	raw := db.DB
	fixtures, err := testfixtures.New(
		testfixtures.Database(raw),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../testdata/fixtures"),
		testfixtures.DangerousSkipTestDatabaseCheck(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := fixtures.Load(); err != nil {
		t.Fatal(err)
	}

	return raw
}
