package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

type Fire struct {
	Name      string
	FireSize  string
	Latitude  string
	Longitude string
	Year      string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func websocketHandler(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	if request.URL.Path == "/concurrent" {
		wg.Add(1)
		go sendToWebSocketConcurrent(conn, &wg)
	} else if request.URL.Path == "/serial" {
		wg.Add(1)
		go sendToWebSocketSerial(conn, &wg)
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "mapPage.html")
	})
	http.HandleFunc("/serial", websocketHandler)
	http.HandleFunc("/concurrent", websocketHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func sendToWebSocketSerial(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()

	fires, err := fetchFireDataSerial()
	if err != nil {
		log.Println("Error fetching data from database:", err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("Serial query took %s\n", duration)

	firesJSON, err := json.Marshal(fires)
	if err != nil {
		log.Println("Error marshaling fires:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, firesJSON)
	if err != nil {
		log.Println("Error sending fires through WebSocket:", err)
		return
	}
	log.Printf("Sent serial data1")

}

func sendToWebSocketConcurrent(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()

	firesCh := make(chan []Fire, 1)

	go func() {
		defer close(firesCh)
		fires, err := fetchFireDataConcurrent()
		if err != nil {
			log.Println("Error fetching data from database concurrently:", err)
			return
		}
		firesCh <- fires
	}()

	fires := <-firesCh

	duration := time.Since(startTime)
	log.Printf("Concurrent query took %s\n", duration)

	firesJSON, err := json.Marshal(fires)
	if err != nil {
		log.Println("Error marshaling fires:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, firesJSON)
	if err != nil {
		log.Println("Error sending fires through WebSocket:", err)
		return
	}
	log.Printf("Sent concurrent data1")

}

func fetchFireDataSerial() ([]Fire, error) {
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		return nil, err
	}
	defer firedb.Close()

	row, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
							FROM Fires LIMIT 1000;
							`)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var fires []Fire

	for row.Next() {
		var fire Fire
		row.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year)
		log.SetFlags(0)
		fires = append(fires, fire)
	}

	return fires, nil
}

func fetchFireDataConcurrent() ([]Fire, error) {
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		return nil, err
	}
	defer firedb.Close()

	row, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
							FROM Fires LIMIT 1000;
							`)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	firesCh := make(chan Fire, 1000)
	done := make(chan struct{})
	var wg sync.WaitGroup

	go func() {
		defer close(firesCh)
		defer close(done)

		for row.Next() {
			var fire Fire
			row.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year)
			log.SetFlags(0)

			wg.Add(1)
			go func(fire Fire) {
				defer wg.Done()
				firesCh <- fire
			}(fire)
		}
		wg.Wait()

	}()

	var fires []Fire

	go func() {
		for fire := range firesCh {
			fires = append(fires, fire)
		}
	}()

	<-done
	//wg.Wait()

	return fires, nil
}
