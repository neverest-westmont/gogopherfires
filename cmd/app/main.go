package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func getRoute(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "mapPage.html")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	// http.HandleFunc("/", getRoute)
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal(err)
	// }

	var wg sync.WaitGroup
	wg.Add(2)
	go callLinear(&wg)
	go callConcurrent(&wg)

	fmt.Println("Waiting for goroutines to finish...")
	wg.Wait()
	fmt.Println("Finished main.")

}

func callLinear(wg *sync.WaitGroup) {
	defer wg.Done()
	defer timer("callLinear")()

	displayDataLinear("linearOut1")
	fmt.Println("Finished callLinear.")
}

func callConcurrent(wg *sync.WaitGroup) {
	defer wg.Done()
	var wgc sync.WaitGroup
	defer timer("callConcurrent")()

	wgc.Add(1)
	go displayDataConcurrent("concurrentOut1", &wgc)
	fmt.Println("Finished callConcurrent.")
}

func displayDataLinear(outfile string) {
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
							ORDER BY FIRE_SIZE ASC LIMIT 3000000`)
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

func displayDataConcurrent(outfile string, wg *sync.WaitGroup) {
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
							ORDER BY FIRE_SIZE ASC LIMIT 3000000`)
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
		duration := time.Since(start)
		fmt.Printf("%s took %10f seconds\n", name, duration.Seconds())
	}
}
