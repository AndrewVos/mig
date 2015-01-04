package mig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var supportedDrivers = map[string]bool{
	"postgres": true,
	"sqlite3":  true,
}

func Migrate(driver string, databaseURL string, migrationsPath string) error {
	if !supportedDrivers[driver] {
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
	var migrations []*Migration
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), ".sql") {
				migrationFiles = append(migrationFiles, file.Name())
			}
		}
	}
	sort.Strings(migrationFiles)

	for _, migrationFile := range migrationFiles {
		migrations = append(migrations, NewMigrationFromPath(path.Join(migrationsPath, migrationFile)))
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
	} else if driver == "sqlite3" {
		_, err := database.Exec(`CREATE TABLE IF NOT EXISTS database_versions(version TEXT);`)
		return err
	}
	return nil
}

type Migration struct {
	Path    string
	Version string
	content string
}

func NewMigrationFromPath(path string) *Migration {
	baseName := filepath.Base(path)
	version := regexp.MustCompile("^\\d+").FindString(baseName)
	return &Migration{
		Path:    path,
		Version: version,
	}
}

func (m Migration) Execute(driver string, database *sqlx.DB) error {
	if driver == "postgres" || driver == "sqlite3" {
		var count int
		err := database.Get(&count, "SELECT COUNT(*) FROM database_versions WHERE version=$1", m.Version)
		if err != nil {
			return m.Error(err)
		}
		if count == 1 {
			return nil
		}
	}
	fmt.Printf("Executing migration %v\n", m.Path)

	contents, err := m.Contents()
	if err != nil {
		return m.Error(err)
	}

	tx, err := database.Begin()
	_, err = tx.Exec(string(contents))
	if err != nil {
		tx.Rollback()
		return m.Error(err)
	}
	_, err = tx.Exec("INSERT INTO database_versions (version) VALUES ($1)", m.Version)
	if err != nil {
		tx.Rollback()
		return m.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return m.Error(err)
	}

	return nil
}

func (m *Migration) Contents() ([]byte, error) {
	b, err := ioutil.ReadFile(m.Path)
	return b, err
}

func (m *Migration) Error(err error) error {
	contents, _ := m.Contents()
	return errors.New(
		fmt.Sprintf("failed to run migration %q\nerror: %v\ncontents: %v\n", m.Path, err, string(contents)),
	)
}
