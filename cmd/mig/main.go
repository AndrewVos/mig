package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

func main() {
	if len(os.Args) > 2 {
		migrationsPath := os.Args[1]
		migrationName := os.Args[2]

		const layout = "20060102150405"
		migrationTime := time.Now().Format(layout)

		migrationName = migrationTime + "_" + migrationName + ".sql"
		fileName := path.Join(migrationsPath, migrationName)

		fmt.Printf("creating migration %v\n", migrationName)
		f, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()

		return
	}
	log.Printf("Usage: mig <migrations_path> <migration_name>\n")
}
