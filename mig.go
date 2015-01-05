package mig

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	MigrationTimeLayout = "20060102150405"
)

func getDriver(driver string) (Driver, error) {
	if driver == "postgres" {
		return &PostgresDriver{}, nil
	} else if driver == "sqlite3" {
		return &SqliteDriver{}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("%q is not a supported driver", driver))
	}
}

func Migrate(driverName string, databaseURL string, migrationsPath string) error {
	driver, err := getDriver(driverName)
	if err != nil {
		return err
	}

	database, err := sqlx.Connect(driverName, databaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	err = driver.CreateVersionsTable(database)
	if err != nil {
		return err
	}

	migrations, err := LoadMigrationsFromPath(migrationsPath)
	if err != nil {
		return err
	}
	return migrations.Up(driver, database)
}

func MigrateDown(driverName string, databaseURL string, migrationsPath string) error {
	driver, err := getDriver(driverName)
	if err != nil {
		return err
	}

	database, err := sqlx.Connect(driverName, databaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	migrations, err := LoadMigrationsFromPath(migrationsPath)
	if err != nil {
		return err
	}
	return migrations.Down(driver, database)
}
