package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var hostName string

func main() {
	rand.Seed(time.Now().UnixNano())
	hostName, _ = os.Hostname()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/random", randomHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println("ERROR starting web server,", err)
	}
}

type RandomResult struct {
	Number int64  `json:"number"`
	Host   string `json:"host"`
}

func randomHandler(w http.ResponseWriter, req *http.Request) {

	result := &RandomResult{
		Number: rand.Int63n(9999),
		Host:   hostName,
	}

	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s", "go away")
}
