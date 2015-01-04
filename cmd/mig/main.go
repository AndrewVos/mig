package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/AndrewVos/mig"
)

func main() {
	if len(os.Args) > 2 {
		migrationsPath := os.Args[1]
		migrationName := os.Args[2]

		migrationTime := time.Now().Format(mig.MigrationTimeLayout)

		migrationName = migrationTime + "_" + migrationName + ".sql"
		fileName := path.Join(migrationsPath, migrationName)

		fmt.Printf("creating migration %v\n", migrationName)
		f, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		_, err = f.Write([]byte("-- up\n\n-- down"))
		if err != nil {
			log.Fatal(err)
		}

		return
	}
	log.Printf("Usage: mig <migrations_path> <migration_name>\n")
}
