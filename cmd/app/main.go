package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go callLinear(&wg)
	go callConcurrent(&wg)

	fmt.Println("Waiting for goroutines to finish...")
	wg.Wait()
	fmt.Println("Finished.")
}

func callLinear(wg *sync.WaitGroup) {
	defer wg.Done()
	var wgl sync.WaitGroup
	defer timer("callLinear")()

	wgl.Add(3)
	displayData1("linearOut1", &wgl)
	displayData2("linearOut2", &wgl)
	displayData3("linearOut3", &wgl)
}

func callConcurrent(wg *sync.WaitGroup) {
	defer wg.Done()
	var wgc sync.WaitGroup
	defer timer("callConcurrent")()

	wgc.Add(3)
	go displayData1("concurrentOut1", &wgc)
	go displayData2("concurrentOut2", &wgc)
	go displayData3("concurrentOut3", &wgc)
}

func displayData1(outfile string, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Create(outfile)
	if err != nil {
		log.Fatal(err)
	}
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()
	row, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
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
		mw := io.MultiWriter(file)
		fmt.Fprintf(mw, "Fire: %s %s %s %s %s\n", name, firesize, latitude, longitude, year)
	}
}

func displayData2(outfile string, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Create(outfile)
	if err != nil {
		log.Fatal(err)
	}
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()
	row, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
							FROM Fires
							WHERE FIRE_SIZE < 22
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
		mw := io.MultiWriter(file)
		fmt.Fprintf(mw, "Fire: %s %s %s %s %s\n", name, firesize, latitude, longitude, year)
	}
}

func displayData3(outfile string, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Create(outfile)
	if err != nil {
		log.Fatal(err)
	}
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()
	row, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
								FROM Fires
								WHERE FIRE_SIZE > 22
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
		mw := io.MultiWriter(file)
		fmt.Fprintf(mw, "Fire: %s %s %s %s %s\n", name, firesize, latitude, longitude, year)
	}
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
