package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	defer timer("main")()
	// database sqlite file should be located in the app folder
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()
	displayData1(firedb)
	displayData2(firedb)
}

func displayData1(db *sql.DB) {
	row, err := db.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
							FROM Fires
							WHERE NWCG_REPORTING_UNIT_NAME = 'Eldorado National Forest'
							AND FIRE_SIZE > 22
							ORDER BY FIRE_SIZE ASC`)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var name string
		var firesize string
		var latitude string
		var longitude string
		var year string
		row.Scan(&name, &firesize, &latitude, &longitude, &year)
		log.SetFlags(0)
		fmt.Printf("Fire: %s %s %s %s %s\n", name, firesize, latitude, longitude, year)
	}
}

func displayData2(db *sql.DB) {
	row, err := db.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
							FROM Fires
							WHERE NWCG_REPORTING_UNIT_NAME = 'Eldorado National Forest'
							AND FIRE_SIZE < 22
							ORDER BY FIRE_SIZE ASC`)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var name string
		var firesize string
		var latitude string
		var longitude string
		var year string
		row.Scan(&name, &firesize, &latitude, &longitude, &year)
		log.SetFlags(0)
		fmt.Printf("Fire: %s %s %s %s %s\n", name, firesize, latitude, longitude, year)
	}
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
