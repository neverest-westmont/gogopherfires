// package main

// import (
// 	"database/sql"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/websocket"

// 	_ "github.com/mattn/go-sqlite3"
// )

// type Fire struct {
// 	Name      string
// 	FireSize  string
// 	Latitude  string
// 	Longitude string
// 	Year      string
// }

// func getRoute(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, "mapPage.html")
// }

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

// func websocketHandler(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	defer conn.Close()

// 	callConcurrent(conn)
// }

// func main() {
// 	// http.HandleFunc("/", getRoute)
// 	// if err := http.ListenAndServe(":8080", nil); err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	http.HandleFunc("/websocket", websocketHandler)

// 	fmt.Println("Finished.")

// 	fmt.Println("Waiting for goroutines to finish...")

// 	callLinear()

// 	fmt.Println("Finished.")

// }

// func callLinear() {
// 	defer timer("callLinear")()
// 	displayLinearData("linearOut")
// }

// func callConcurrent(conn *websocket.Conn) {
// 	defer timer("callConcurrent")()

// 	displayConcurrentData(conn)
// }

// func displayLinearData(outfile string) {
// 	file, err := os.Create(outfile)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer firedb.Close()
// 	rows, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
// 								FROM Fires
// 								`)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for rows.Next() {
// 		var fire Fire
// 		rows.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year)
// 		mw := io.MultiWriter(file)
// 		fmt.Fprintf(mw, "Fire: %s %s %s %s %s\n", fire.Name, fire.FireSize, fire.Latitude, fire.Longitude, fire.Year)
// 	}
// 	fmt.Println("Linear finished")
// }

// func displayConcurrentData(conn *websocket.Conn) {
// 	firedb, err := sql.Open("sqlite3", "../../internal/db/FPA_FOD_20221014.sqlite")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer firedb.Close()

// 	rows, err := firedb.Query(`SELECT FIRE_NAME, FIRE_SIZE, LATITUDE, LONGITUDE, FIRE_YEAR
// 								FROM Fires
// 								`)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()

// 	var wg sync.WaitGroup
// 	jobs := make(chan []string, 200)
// 	results := make(chan string, 200)

// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go worker(jobs, results, &wg)
// 	}

// 	go func() {
// 		for rows.Next() {
// 			var fire Fire
// 			rows.Scan(&fire.Name, &fire.FireSize, &fire.Latitude, &fire.Longitude, &fire.Year)
// 			data := fmt.Sprintf("Fire: %s %s %s %s %s\n", fire.Name, fire.FireSize, fire.Latitude, fire.Longitude, fire.Year)
// 			jobs <- []string{data}
// 		}
// 		close(jobs)
// 	}()

// 	go func() {
// 		wg.Wait()
// 		close(results)
// 	}()

// 	for result := range results {
// 		err := conn.WriteMessage(websocket.TextMessage, []byte(result))
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 	}

// 	fmt.Println("Concurrent finished")
// }

// func worker(jobs <-chan []string, results chan<- string, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	for job := range jobs {
// 		results <- job[0]
// 	}
// }

// func timer(name string) func() {
// 	start := time.Now()
// 	return func() {
// 		duration := time.Since(start)
// 		fmt.Printf("%s took %10f seconds\n", name, duration.Seconds())
// 	}
// }
