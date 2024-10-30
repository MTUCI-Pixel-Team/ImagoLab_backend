package main

import (
	"RestAPI/db"
	"flag"
)

func main() {
	rollbaсk := flag.Bool("rollback", false, "Rollback migrations")
	version := flag.Int("version", -1, "Version of migration to rollback")
	flag.Parse()
	db.RunMigrations(*rollbaсk, *version)
}
