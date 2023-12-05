package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Fire struct {
	Name      string
	FireSize  string
	Latitude  string
	Longitude string
	Year      string
}

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

	channelConcurrent := make(chan []byte)
	channelLinear := make(chan []byte)
	go callConcurrent(channelConcurrent)
	go callLinear(channelLinear)

	fmt.Println("Waiting for goroutines to finish...")

	linChan := <-channelLinear
	concChan := <-channelConcurrent
	fmt.Println(concChan)
	fmt.Println(linChan)

	fmt.Println("Finished.")
}

func callLinear(channelLinear chan []byte) {
	defer timer("callLinear")()
	displayDataLinear("linearOut1", channelLinear)
}

func callConcurrent(channelConcurrent chan []byte) {
	defer timer("callConcurrent")()

	go displayDataConcurrent("concurrentOut1", channelConcurrent)
}

func displayDataLinear(outfile string, channelLinear chan []byte) []byte {

	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()
	row, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
							FROM Fires
							ORDER BY FIRE_SIZE ASC LIMIT 3`)
	if err != nil {
		log.Fatal(err)
	}

	defer row.Close()

	var fires []Fire

	for row.Next() {
		var fire Fire
		row.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year)
		fires = append(fires, fire)
	}
	firesJSON, err := json.Marshal(fires)
	channelLinear <- firesJSON
	return firesJSON
}

func displayDataConcurrent(outfile string, channelConcurrent chan []byte) []byte {

	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()
	row, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
							FROM Fires
							ORDER BY FIRE_SIZE ASC LIMIT 3`)
	if err != nil {
		log.Fatal(err)
	}

	defer row.Close()

	var fires []Fire

	for row.Next() {
		var fire Fire
		row.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year)
		fires = append(fires, fire)
	}
	firesJSON, err := json.Marshal(fires)
	channelConcurrent <- firesJSON
	return firesJSON
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		fmt.Printf("%s took %10f seconds\n", name, duration.Seconds())
	}
}
