package main

import (
	"database/sql"
	"github.com/mattn/go-sqlite3"
	"log"
)

func main() {
	db, err := sql.Open("sqlite3", "FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

}
