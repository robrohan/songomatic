package repository

import (
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

// OpenDatabase Open up the database connection
func OpenDatabase(driver string, connection string, schema string) (*sqlx.DB, error) {
	conn := strings.ReplaceAll(connection, "{schema}", schema)
	db, err := sqlx.Open(driver, conn)
	if err != nil {
		log.Printf("Failed to open data store")
		return nil, err
	}

	// Run database migrations
	err = UpdateDataStore(driver, db)
	if err != nil {
		log.Printf("Failed to upgrade the data store")
		return nil, err
	}

	return db, nil
}

// UpdateDataStore Run any migrations that need to run
func UpdateDataStore(driver string, db *sqlx.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}
	n, err := migrate.Exec(db.DB, driver, migrations, migrate.Up)
	if err != nil {
		log.Printf("Filed migrations\n")
		return err
	}
	log.Printf("Applied %d migrations\n", n)
	return nil
}
