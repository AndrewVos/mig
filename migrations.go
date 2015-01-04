package mig

import (
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Migrations []*Migration

func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Less(i, j int) bool { return m[i].Version.Before(m[j].Version) }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

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

func (m Migrations) Up(driver Driver, database *sqlx.DB) error {
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

func (m Migrations) Down(driver Driver, database *sqlx.DB) error {
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
