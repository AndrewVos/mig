package mig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
)

type Migrations []*Migration

func (m Migrations) ExecuteInOrder(database *sqlx.DB) error {
	sort.Sort(m)
	for _, migration := range m {
		err := migration.Execute(database)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Less(i, j int) bool { return m[i].Version.Before(m[j].Version) }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

type Migration struct {
	Path    string
	Version time.Time
}

func NewMigrationFromPath(path string) (*Migration, error) {
	baseName := filepath.Base(path)
	unparsedVersion := regexp.MustCompile("^\\d+").FindString(baseName)
	version, err := time.Parse(MigrationTimeLayout, unparsedVersion)
	if err != nil {
		return nil, err
	}
	return &Migration{
		Path:    path,
		Version: version,
	}, nil
}

func (m Migration) Execute(database *sqlx.DB) error {
	versionHasBeenExecuted, err := currentDriver.VersionHasBeenExecuted(database, m.Version)
	if err != nil {
		return err
	}
	if versionHasBeenExecuted {
		return nil
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
	err = currentDriver.MarkVersionAsExecuted(tx, m.Version)
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
