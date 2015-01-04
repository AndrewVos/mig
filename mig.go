package mig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/jmoiron/sqlx"
)

var currentDriver Driver

const (
	MigrationTimeLayout = "20060102150405"
)

func Migrate(driver string, databaseURL string, migrationsPath string) error {
	if driver == "postgres" {
		currentDriver = &PostgresDriver{}
	} else if driver == "sqlite3" {
		currentDriver = &SqliteDriver{}
	} else {
		return errors.New(fmt.Sprintf("%q is not a supported driver", driver))
	}

	database, err := sqlx.Connect(driver, databaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	err = currentDriver.CreateVersionsTable(database)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return err
	}

	var migrations Migrations
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), ".sql") {
				migration, err := NewMigrationFromPath(path.Join(migrationsPath, file.Name()))
				if err != nil {
					return err
				}
				migrations = append(migrations, migration)
			}
		}
	}

	return migrations.ExecuteInOrder(database)
}
