# mig

## Installing the command line tool

```
go get github.com/AndrewVos/mig/cmd/mig
```

## Creating new migrations

```
$ mkdir migrations
$ mig migrations create_table
creating migration 20150104013210_create_table.sql
$ mig migrations add_column_to_table
creating migration 20150104013218_add_column_to_table.sql
```

## Running migrations

```golang

import (
	"github.com/AndrewVos/mig"
	"log"
)

func main() {
	err := mig.Migrate("postgres", databaseURL(), "./migrations")
	if err != nil {
		log.Fatal(err)
	}
}
```
