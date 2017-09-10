package main

import (
	"log"
	"os"

	"github.com/jorinvo/slangbrain/brain"
)

const cliUsage = `Slangbrain Stat

Usage: %s [path_to_db_file]

Slangbrain Stat outputs text-based statistics for a passed Slangbrain DB file to stdout.
It is best run with a DB backup as file to not put more load on Slangbrain itself
and it also ensures that backups are working properly.
`

func main() {
	errs := log.New(os.Stderr, "", 0)

	if len(os.Args) < 2 {
		errs.Fatalf(cliUsage, os.Args[0])
	}
	db := os.Args[1]
	if db == "" || db == "-h" || db == "-help" || db == "--help" || db == "help" {
		errs.Fatalf(cliUsage, os.Args[0])
	}

	if _, err := os.Stat(db); os.IsNotExist(err) {
		errs.Fatalf("no file found at '%s'", db)
	}

	// Setup database
	store, err := brain.New(db)
	if err != nil {
		errs.Fatalln("failed to create store:", err)
	}
	defer func() {
		if err = store.Close(); err != nil {
			errs.Println("failed to close store:", err)
		}
	}()

	if err := store.WriteStat(os.Stdout); err != nil {
		errs.Fatalln("failed to write stat:", err)
	}
}
