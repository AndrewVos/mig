# mig

## Creating new migrations

Install the command line tool:

```
go get github.com/AndrewVos/mig/cmd/mig
```

```
$ mkdir migrations
$ mig ./migrations create_table
creating migration 20150104013210_create_table.sql
$ mig ./migrations add_column_to_table
creating migration 20150104013218_add_column_to_table.sql
```

## Running migrations

First install the client library:

```
go get github.com/AndrewVos/mig
```

```golang
import (
	"github.com/AndrewVos/mig"
	"log"
)

func main() {
	// err := mig.Migrate("sqlite3", "file.sqlite", "./migrations")
	err := mig.Migrate("postgres", "host=/var/run/postgresql dbname=my_database sslmode=disable", "./migrations")
	if err != nil {
		log.Fatal(err)
	}
}
```

## Supported drivers

- postgres
- sqlite3
