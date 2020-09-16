package db

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"

	"u9k/db/migrate/cockroachdb"
)

func applyMigrations(db *pgx.Conn, forceVersion int) error {
	driver, err := cockroachdb.WithInstance(db, new(cockroachdb.Config))
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations/", "cockroachdb", driver)
	if err != nil {
		return err
	}

	printMigrationVersion(m)
	if forceVersion > 0 {
		log.Printf("Forcing migration version to %d\n", forceVersion)
		err = m.Force(forceVersion)
	} else {
		err = m.Up()
	}
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	if err == migrate.ErrNoChange {
		log.Printf("No migrations need to be applied\n")
	} else {
		printMigrationVersion(m)
	}

	source_err, db_err := m.Close()
	if source_err != nil {
		return source_err
	}
	if db_err != nil {
		return db_err
	}

	return nil
}

func printMigrationVersion(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			log.Printf("No migrations found in database")
		} else {
			log.Fatalf("Error checking migration version: %s\n", err)
		}
	} else {
		log.Printf("Current migration version: %d (dirty: %t)\n", version, dirty)
	}
}
