package mig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func cleanup() {
	os.Remove("./db.sqlite")
	os.RemoveAll("./migrations")
}

var currentMigrationTime = time.Now()

func createMigration(t *testing.T, migrationName string, sql string) {
	os.MkdirAll("migrations", 0700)

	currentMigrationTime = currentMigrationTime.Add(1 * time.Second)

	const layout = "20060102150405"
	migrationTime := currentMigrationTime.Format(layout)

	migrationName = migrationTime + "_" + migrationName + ".sql"
	fileName := path.Join("migrations", migrationName)

	err := ioutil.WriteFile(fileName, []byte(sql), 0700)
	if err != nil {
		t.Fatal(err)
	}
}

func assertValueOfFirstThing(t *testing.T, expectedValue string) {
	db, err := sqlx.Connect("sqlite3", "db.sqlite")
	if err != nil {
		t.Fatal(err)
	}

	var value string
	err = db.Get(&value, "SELECT value FROM things LIMIT 1")
	if err != nil {
		t.Fatal(err)
	}

	if value != expectedValue {
		t.Errorf("Expected value to be %q but was %q", expectedValue, value)
	}
}

func runMigrations(t *testing.T) {
	err := Migrate("sqlite3", "db.sqlite", "migrations")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunsMigrationsInOrder(t *testing.T) {
	defer cleanup()

	createMigration(t, "create_things", "CREATE TABLE things(value TEXT);")
	createMigration(t, "populate_things", "INSERT INTO things (value) VALUES ('hello:');")
	for i := 0; i < 5; i++ {
		createMigration(t, "concat_values", fmt.Sprintf("UPDATE things SET value=value || '%v';", i))
	}

	runMigrations(t)
	assertValueOfFirstThing(t, "hello:01234")
}

func TestMigrationsDontRunTwice(t *testing.T) {
	defer cleanup()
	db, err := sqlx.Connect("sqlite3", "db.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE things(value TEXT);
		INSERT INTO things (value) VALUES ('hello:');
	`)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		createMigration(t, "concat_values", fmt.Sprintf("UPDATE things SET value=value || '%v';", i))
	}

	runMigrations(t)
	runMigrations(t)

	assertValueOfFirstThing(t, "hello:01234")
}
