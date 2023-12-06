package main

import (
	"database/sql"
	"encoding/json"
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
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

func sendToSerialWebSocket(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	fires, err := fetchSerialFireData()
	if err != nil {
		log.Println("Error fetching data from database:", err)
		return
	}

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
}

func sendToConcurrentWebSocket(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	fires, err := fetchConcurrentFireData()
	if err != nil {
		log.Println("Error fetching data from database:", err)
		return
	}

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
}

func fetchSerialFireData() ([]Fire, error) {
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		return nil, err
	}
	defer firedb.Close()

	query, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
							FROM Fires
							LIMIT 1000`)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	var fires []Fire

	for query.Next() {
		var fire Fire
		query.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year)
		log.SetFlags(0)
		fires = append(fires, fire)
	}

	return fires, nil
}

func fetchConcurrentFireData() ([]Fire, error) {
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()

	query, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
								FROM Fires
								LIMIT 1000
								`)
	if err != nil {
		log.Fatal(err)
	}
	defer query.Close()

	var wg sync.WaitGroup
	jobs := make(chan []Fire, 200)
	results := make(chan Fire, 200)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	var fires []Fire

	go func() {
		for query.Next() {
			var fire Fire
			query.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year)
			data := []Fire{fire}
			jobs <- data
			fires = append(fires, fire)
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		wg.Wait()
	}()

	var resultFires []Fire

	for result := range results {
		resultFires = append(resultFires, result)
	}

	return resultFires, nil
}

func worker(jobs <-chan []Fire, results chan<- Fire, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		results <- job[0]
	}
}
