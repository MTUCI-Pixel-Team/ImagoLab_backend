package main

import (
	"RestAPI/db"
	"flag"
)

func main() {
	rollbak := flag.Bool("rollback", false, "Rollback migrations")
	flag.Parse()
	db.RunMigrations(*rollbak)
}
