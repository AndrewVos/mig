package mig

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"io/ioutil"
	"path"
	"regexp"
	"sort"
	"strings"
)

func Migrate(driver string, databaseURL string, migrationsPath string) error {
	if driver != "postgres" {
		return errors.New(fmt.Sprintf("Driver %q is not supported", driver))
	}

	database, err := sqlx.Connect(driver, databaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	err = createVersionsTable(driver, database)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return err
	}

	var migrationFiles []string
	var migrations []Migration
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), ".sql") {
				migrationFiles = append(migrationFiles, file.Name())
			}
		}
	}
	sort.Strings(migrationFiles)

	for _, migrationFile := range migrationFiles {
		version := regexp.MustCompile("^\\d+").FindString(migrationFile)
		if err != nil {
			return err
		}
		migrations = append(migrations, Migration{
			Path:    path.Join(migrationsPath, migrationFile),
			Version: version,
		})
	}

	for _, migration := range migrations {
		err := migration.Execute(driver, database)
		if err != nil {
			return err
		}
	}

	return nil
}

func createVersionsTable(driver string, database *sqlx.DB) error {
	if driver == "postgres" {
		_, err := database.Exec(`CREATE TABLE IF NOT EXISTS database_versions(version TEXT);`)
		return err
	}
	return nil
}

type Migration struct {
	Path    string
	Version string
}

func (m Migration) Execute(driver string, database *sqlx.DB) error {
	if driver == "postgres" {
		var count int
		err := database.Get(&count, "SELECT COUNT(*) FROM database_versions WHERE version=$1", m.Version)
		if err != nil {
			return migrationError(m.Path, err)
		}
		if count == 1 {
			return nil
		}
	}
	fmt.Printf("Executing migration %v\n", m.Path)

	b, err := ioutil.ReadFile(m.Path)
	if err != nil {
		return migrationError(m.Path, err)
	}

	tx, err := database.Begin()
	_, err = tx.Exec(string(b))
	if err != nil {
		tx.Rollback()
		return migrationError(m.Path, err)
	}
	_, err = tx.Exec("INSERT INTO database_versions (version) VALUES ($1)", m.Version)
	if err != nil {
		tx.Rollback()
		return migrationError(m.Path, err)
	}

	err = tx.Commit()
	if err != nil {
		return migrationError(m.Path, err)
	}

	return nil
}

func migrationError(migration string, err error) error {
	return errors.New(
		fmt.Sprintf("failed to run migration %q\n%v", migration, err),
	)
}
