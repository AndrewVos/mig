package mig

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Migration struct {
	Name     string
	Path     string
	Version  time.Time
	Contents struct {
		Up   string
		Down string
	}
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
		return m.upError(err)
	}
	err = driver.MarkVersionAsExecuted(tx, m.VersionAsString())
	if err != nil {
		tx.Rollback()
		return m.upError(err)
	}

	err = tx.Commit()
	if err != nil {
		return m.upError(err)
	}

	return nil
}

func (m Migration) Down(driver Driver, database *sqlx.DB) error {
	tx, err := database.Begin()
	_, err = tx.Exec(m.Contents.Down)
	if err != nil {
		tx.Rollback()
		return m.downError(err)
	}
	err = driver.UnmarkVersionAsExecuted(tx, m.VersionAsString())
	if err != nil {
		tx.Rollback()
		return m.downError(err)
	}

	err = tx.Commit()
	if err != nil {
		return m.downError(err)
	}

	return nil
}

func (m *Migration) VersionAsString() string {
	return m.Version.Format(MigrationTimeLayout)
}

func (m *Migration) upError(err error) error {
	return errors.New(
		fmt.Sprintf("failed to run migration %q\nerror: %v\ncontents: %v\n", m.Path, err, m.Contents.Up),
	)
}

func (m *Migration) downError(err error) error {
	return errors.New(
		fmt.Sprintf("failed to rollback migration %q\nerror: %v\ncontents: %v\n", m.Path, err, m.Contents.Down),
	)
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
