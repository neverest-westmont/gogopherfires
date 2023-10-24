package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "../FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(db)

	defer db.Close()

}
