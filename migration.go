package mig

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Migrations []*Migration

func LoadMigrationsFromPath(migrationsPath string) (Migrations, error) {
	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}

	var migrations Migrations
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), ".sql") {
				migration, err := NewMigrationFromPath(path.Join(migrationsPath, file.Name()))
				if err != nil {
					return nil, err
				}
				migrations = append(migrations, migration)
			}
		}
	}

	return migrations, nil
}

func (m Migrations) ExecuteInOrder(driver Driver, database *sqlx.DB) error {
	sort.Sort(m)
	for _, migration := range m {
		hasBeenMigrated, err := migration.HasBeenMigrated(driver, database)
		if err != nil {
			return err
		}
		if !hasBeenMigrated {
			fmt.Printf("[up] %v\n%v\n", migration.Name, migration.Contents.Up)
			err := migration.Up(driver, database)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m Migrations) Rollback(driver Driver, database *sqlx.DB) error {
	sort.Reverse(m)
	for _, migration := range m {
		hasBeenMigrated, err := migration.HasBeenMigrated(driver, database)
		if err != nil {
			return err
		}
		if hasBeenMigrated {
			fmt.Printf("[down] %v\n%v\n", migration.Name, migration.Contents.Down)
			err := migration.Down(driver, database)
			return err
		}
	}
	return nil
}

func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Less(i, j int) bool { return m[i].Version.Before(m[j].Version) }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

type Migration struct {
	Name     string
	Path     string
	Version  time.Time
	Contents MigrationContents
}

func NewMigrationFromPath(path string) (*Migration, error) {
	baseName := filepath.Base(path)
	unparsedVersion := regexp.MustCompile("^\\d+").FindString(baseName)
	version, err := time.Parse(MigrationTimeLayout, unparsedVersion)
	if err != nil {
		return nil, err
	}

	migration := &Migration{
		Name:    baseName,
		Path:    path,
		Version: version,
	}

	err = migration.readContents()
	return migration, err
}

func (m Migration) HasBeenMigrated(driver Driver, database *sqlx.DB) (bool, error) {
	return driver.VersionHasBeenExecuted(database, m.VersionAsString())
}

func (m Migration) Up(driver Driver, database *sqlx.DB) error {
	tx, err := database.Begin()
	_, err = tx.Exec(m.Contents.Up)
	if err != nil {
		tx.Rollback()
		return m.UpError(err)
	}
	err = driver.MarkVersionAsExecuted(tx, m.VersionAsString())
	if err != nil {
		tx.Rollback()
		return m.UpError(err)
	}

	err = tx.Commit()
	if err != nil {
		return m.UpError(err)
	}

	return nil
}

func (m Migration) Down(driver Driver, database *sqlx.DB) error {
	tx, err := database.Begin()
	_, err = tx.Exec(m.Contents.Down)
	if err != nil {
		tx.Rollback()
		return m.DownError(err)
	}
	err = driver.UnmarkVersionAsExecuted(tx, m.VersionAsString())
	if err != nil {
		tx.Rollback()
		return m.DownError(err)
	}

	err = tx.Commit()
	if err != nil {
		return m.DownError(err)
	}

	return nil
}

type MigrationContents struct {
	Up   string
	Down string
}

func (m *Migration) readContents() error {
	file, err := os.Open(m.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	up := true
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "-- up") {
			up = true
		} else if strings.HasPrefix(strings.TrimSpace(line), "-- down") {
			up = false
		} else {
			if up {
				m.Contents.Up += line + "\n"
			} else if !up {
				m.Contents.Down += line + "\n"
			}
		}
	}
	return nil
}

func (m *Migration) VersionAsString() string {
	return m.Version.Format(MigrationTimeLayout)
}

func (m *Migration) UpError(err error) error {
	return errors.New(
		fmt.Sprintf("failed to run migration %q\nerror: %v\ncontents: %v\n", m.Path, err, m.Contents.Up),
	)
}

func (m *Migration) DownError(err error) error {
	return errors.New(
		fmt.Sprintf("failed to rollback migration %q\nerror: %v\ncontents: %v\n", m.Path, err, m.Contents.Down),
	)
}
