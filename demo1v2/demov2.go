package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	appVersion         = "v2"
	maxOngoingRequests = 100
	maxRequestsServed  = 40000
)

var (
	ongoingRequests int32
	totalServed     int32
)

type simpleResponse struct {
	ReqTime    string `json:"reqTime"`
	AppVersion string `json:"appVersion"`
}

func main() {
	log.SetOutput(os.Stdout)
	srv := &http.Server{Addr: ":8080"}
	http.HandleFunc("/", rootHandler)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Httpserver: ListenAndServe() exit error: %s", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)

	<-sigCh
	log.Println("Received TERM signal...")
	if err := srv.Shutdown(context.Background()); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
	log.Println("Exited.")

	os.Exit(0)
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	if atomic.LoadInt32(&totalServed) > maxRequestsServed {
		http.Error(w, "server is no longer serving requests, please go away", http.StatusServiceUnavailable)
		atomic.AddInt32(&ongoingRequests, -1)
		return
	}

	if atomic.AddInt32(&ongoingRequests, 1) > maxOngoingRequests {
		http.Error(w, "server is too busy, try again later", http.StatusServiceUnavailable)
		atomic.AddInt32(&ongoingRequests, -1)
		return
	}

	startTs := time.Now()

	// sleep between 50 and 150 milliseconds
	randomizeSleep(50, 150)

	responseData := &simpleResponse{time.Since(startTs).String(), appVersion}

	jsonData, err := json.Marshal(responseData)
	if err != nil {
		atomic.AddInt32(&ongoingRequests, -1)
		http.Error(w, "internal error: failed to marshal json"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonData)
	atomic.AddInt32(&ongoingRequests, -1)
	atomic.AddInt32(&totalServed, 1)
}

func randomizeSleep(min, max int) {
	rand.Seed(time.Now().UnixNano())
	interval := rand.Intn(max-min) + min
	time.Sleep(time.Millisecond * time.Duration(interval))
}
