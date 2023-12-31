/*
 * Westmont College Fall 2023 CS 105 Programming Languages
 * Group Project: Golang Fire Map
 *
 * Created by:
 * Trevor English (tenglish@westmont.edu)
 * Nancy Everest (neverest@westmont.edu)
 * Allie Peterson (alpeterson@westmont.edu)
 *
 */

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

type Fire struct {
	Name      string
	FireSize  string
	Latitude  string
	Longitude string
	Year      string
	Forest    string
	Cause     string
	County    string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "mapPage.html")
	})
	http.HandleFunc("/serial", serialWebsocketHandler)
	http.HandleFunc("/concurrent", concurrentWebsocketHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func serialWebsocketHandler(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go sendToSerialWebSocket(conn, &wg)
}

func concurrentWebsocketHandler(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go sendToConcurrentWebSocket(conn, &wg)
}

func sendToSerialWebSocket(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	fires, err := fetchSerialFireData()
	if err != nil {
		log.Println("Error fetching data serially from database:", err)
		return
	}

	firesJSON, err := json.Marshal(fires)
	if err != nil {
		log.Println("Error marshaling serial fires:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, firesJSON)
	if err != nil {
		log.Println("Error sending serial fires through WebSocket:", err)
		return
	}
	fmt.Println("Websocket successfully sent serial data")

}

func sendToConcurrentWebSocket(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	fires, err := fetchConcurrentFireData()
	if err != nil {
		log.Println("Error fetching data concurrently from database:", err)
		return
	}

	firesJSON, err := json.Marshal(fires)
	if err != nil {
		log.Println("Error marshaling concurrent fires:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, firesJSON)
	if err != nil {
		log.Println("Error sending concurrent fires through WebSocket:", err)
		return
	}
	fmt.Println("Websocket successfully sent concurrent data")
}

func fetchSerialFireData() ([]Fire, error) {
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")

	if err != nil {
		return nil, err
	}
	defer firedb.Close()

	query, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR, NWCG_REPORTING_UNIT_NAME, NWCG_GENERAL_CAUSE, FIPS_NAME
							FROM Fires
							LIMIT 1000`)
	if err != nil {
		return nil, err
	} else {
		fmt.Println("Serial query successful")
	}

	defer query.Close()

	var fires []Fire

	for query.Next() {
		var fire Fire
		query.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year, &fire.Forest, &fire.Cause, &fire.County)
		log.SetFlags(0)
		fires = append(fires, fire)
	}

	return fires, nil
}

func fetchConcurrentFireData() ([]Fire, error) {
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer firedb.Close()

	query, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR, NWCG_REPORTING_UNIT_NAME, NWCG_GENERAL_CAUSE, FIPS_NAME
								FROM Fires
								LIMIT 1000
								`)
	if err != nil {
		return nil, err
	} else {
		fmt.Println("Concurrent query successful")
	}

	defer query.Close()

	var wg sync.WaitGroup
	jobs := make(chan []Fire)
	results := make(chan Fire)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	go func() {
		for query.Next() {
			var fire Fire
			query.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year, &fire.Forest, &fire.Cause, &fire.County)
			data := []Fire{fire}
			jobs <- data

		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var fires []Fire

	for result := range results {
		fires = append(fires, result)
	}

	return fires, nil
}

func worker(jobs <-chan []Fire, results chan<- Fire, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		results <- job[0]
	}
}
