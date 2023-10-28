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
	row, err := db.Query(`SELECT NWCG_REPORTING_UNIT_NAME, FIRE_SIZE, LATITUDE, LONGITUDE
							FROM Fires
							WHERE NWCG_REPORTING_UNIT_NAME = 'Eldorado National Forest'
							AND FIRE_SIZE > 22
							ORDER BY FIRE_SIZE DESC`)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var name string
		var firesize string
		var latitude string
		var longitude string
		row.Scan(&name, &firesize, &latitude, &longitude)
		log.Println("Fire: ", name, " ", firesize, " ", latitude, " ", longitude)
	}
}
