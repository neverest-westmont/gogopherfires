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

	fmt.Println("Waiting for goroutines to finish...")

	callConcurrent()
	callLinear()

	fmt.Println("Finished.")

}

func callLinear() {
	defer timer("callLinear")()
	displayLinearData("linearOut")
}

func callConcurrent() {
	defer timer("callConcurrent")()

	displayConcurrentData("concurrentOut")

}

func displayLinearData(outfile string) {
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
								`)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
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
	fmt.Println("Linear finished")
}

func displayConcurrentData(outfile string) {
	file, err := os.Create(outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()

	rows, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
								FROM Fires
							`)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	dataChannel := make(chan string, 100)

	mw := io.MultiWriter(file)

	// I arbitrarilly chose 300 so that it does 300 queries
	for i := 0; i < 300; i++ {
		wg.Add(1)
		go concurrentQuery(rows, dataChannel, &wg, mw)
	}

	go func() {
		wg.Wait()
		close(dataChannel)
	}()

	fmt.Println("Concurrent finished")
}

func concurrentQuery(rows *sql.Rows, dataChannel chan<- string, wg *sync.WaitGroup, mw io.Writer) {
	defer wg.Done()
	for rows.Next() {
		var name string
		var firesize string
		var latitude string
		var longitude string
		var year string

		err := rows.Scan(&name, &firesize, &latitude, &longitude, &year)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintf(mw, "Fire: %s %s %s %s %s\n", name, firesize, latitude, longitude, year)
		dataChannel <- fmt.Sprintf("Fire: %s %s %s %s %s\n", name, firesize, latitude, longitude, year)

	}
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		fmt.Printf("%s took %10f seconds\n", name, duration.Seconds())
	}
}
