package mig

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Driver interface {
	CreateVersionsTable(database *sqlx.DB) error
	VersionHasBeenExecuted(database *sqlx.DB, version time.Time) (bool, error)
	MarkVersionAsExecuted(transaction *sql.Tx, version time.Time) error
}

type PostgresDriver struct{}

func (d *PostgresDriver) CreateVersionsTable(database *sqlx.DB) error {
	_, err := database.Exec(`CREATE TABLE IF NOT EXISTS database_versions(version TEXT);`)
	return err
}

func (d *PostgresDriver) VersionHasBeenExecuted(database *sqlx.DB, version time.Time) (bool, error) {
	var count int
	err := database.Get(&count, "SELECT COUNT(*) FROM database_versions WHERE version=$1", version)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *PostgresDriver) MarkVersionAsExecuted(transaction *sql.Tx, version time.Time) error {
	_, err := transaction.Exec("INSERT INTO database_versions (version) VALUES ($1)", version)
	return err
}

type SqliteDriver struct {
	PostgresDriver
}
