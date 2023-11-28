package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

func getRoute(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "mapPage.html")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go callLinear(conn, &wg)
	//go callConcurrent(conn, &wg)

	wg.Wait()
}

func main() {
	// http.HandleFunc("/", getRoute)
	http.HandleFunc("/websocket", websocketHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	fmt.Println("Waiting for goroutines to finish...")
	wg.Wait()
	fmt.Println("Finished.")
}

func callLinear(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	var wgl sync.WaitGroup
	defer timer("callLinear")()

	wgl.Add(1)
	linearFires := displayData("linearOut1", &wgl)
	linearFiresJSON, err := json.Marshal(linearFires)
	if err != nil {
		log.Println("Error marshaling linearFires:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, linearFiresJSON)
	if err != nil {
		log.Println("Error sending linearFires through WebSocket:", err)
		return
	}
}

func callConcurrent(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	var wgc sync.WaitGroup
	defer timer("callConcurrent")()

	wgc.Add(1)
	concurrentFires := displayData("concurrentOut1", &wgc)
	concurrentFiresJSON, err := json.Marshal(concurrentFires)
	if err != nil {
		log.Println("Error marshaling linearFires:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, concurrentFiresJSON)
	if err != nil {
		log.Println("Error sending linearFires through WebSocket:", err)
		return
	}
}

func displayData(outfile string, wg *sync.WaitGroup) []Fire {
	defer wg.Done()
	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer firedb.Close()
	row, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
							FROM Fires
							WHERE NWCG_REPORTING_UNIT_NAME = 'Eldorado National Forest'
							ORDER BY FIRE_SIZE ASC LIMIT 100`)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var fires []Fire

	for row.Next() {
		var fire Fire
		row.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year)
		log.SetFlags(0)
		fmt.Println(fire)
		fires = append(fires, fire)
	}
	fmt.Println(fires)

	return fires
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		fmt.Printf("%s took %10f seconds\n", name, duration.Seconds())
	}
}
