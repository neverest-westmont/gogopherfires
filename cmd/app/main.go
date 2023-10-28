package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// database sqlite file should be located in the app folder
	firedb, err := sql.Open("sqlite3", "../FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()
	displayData(firedb)
}

func displayData(db *sql.DB) {
	row, err := db.Query("SELECT NWCG_REPORTING_UNIT_NAME, FIRE_SIZE FROM Fires WHERE NWCG_REPORTING_UNIT_NAME = 'Eldorado National Forest' ORDER BY FIRE_SIZE")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var name string
		var firesize string
		row.Scan(&name, &firesize)
		log.Println("Fire: ", name, " ", firesize)
	}
}
