# mig

## Installing the command line tool

```
go get github.com/AndrewVos/mig/cmd
```

## Creating a new migration file

```
mkdir migrations
mig migrations create_table
mig migrations add_column_to_table
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
