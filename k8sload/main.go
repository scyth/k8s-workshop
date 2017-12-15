package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	influxc "github.com/influxdata/influxdb/client/v2"
)

const (
	InfluxDbName = "loadtest"
)

var (
	maxConn        int
	remoteHostPort string
)

func init() {
	flag.IntVar(&maxConn, "maxconn", 1, "max number of parallel connections")
	flag.StringVar(&remoteHostPort, "target", "localhost:8080", "remote target host:port")
	flag.Parse()
}

type webResponse struct {
	ReqTime    string `json:"reqTime"`
	AppVersion string `json:"appVersion"`
	Error      error
}

type results struct {
	success int64
	failure int64
	numV1   int64
	numV2   int64
	total   int64
}

func (r *results) clear() {
	r.failure, r.success, r.numV1, r.numV2, r.total = 0, 0, 0, 0, 0
}

func (r *results) add(msg *webResponse) {
	if msg.Error == nil {
		r.success += 1
	} else {
		r.failure += 1
	}

	if msg.AppVersion == "v1" {
		r.numV1 += 1
	} else if msg.AppVersion == "v2" {
		r.numV2 += 1
	}
	r.total += 1
}

func (r *results) String() string {
	return fmt.Sprintf("numRequests: %d\tnumVer1 %d\tnumVer2 %d\tnumSucceeded: %d\tnumFailed %d", r.total, r.numV1, r.numV2, r.success, r.failure)
}

type client struct {
	abortCh    chan bool
	responseCh chan *webResponse
	wg         *sync.WaitGroup
	url        string
}

func (c *client) run() {
	for {
		select {
		case <-c.abortCh:
			c.wg.Done()
			return
		default:
			c.makeRequest()
		}
	}
}

func (c *client) makeRequest() {
	resp, err := http.Get(c.url)
	if err != nil {
		log.Println("ERROR: makeRequest() failed,", err)
		time.Sleep(time.Millisecond * 150)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ERROR: makeRerquest() failed to read response,", err)
		return
	}

	if resp.StatusCode != 200 {
		c.responseCh <- &webResponse{Error: errors.New(string(body))}
		time.Sleep(time.Millisecond * 150)
		return
	}

	var jsonResponse webResponse
	if err = json.Unmarshal(body, &jsonResponse); err != nil {
		log.Println("ERROR: makeRequest() failed to parse json response,", err)
		return
	}
	c.responseCh <- &jsonResponse
}

func clientFactory(abortCh chan bool, output chan *webResponse) chan bool {
	doneCh := make(chan bool, 1)
	go func(doneCh, abortCh chan bool, outputCh chan *webResponse) {

		var wg sync.WaitGroup

		// initialize clients
		clients := make([]*client, maxConn)
		url := "http://" + remoteHostPort + "/"
		for i := 0; i < maxConn; i++ {
			clients[i] = &client{make(chan bool), output, &wg, url}
			wg.Add(1)
			go clients[i].run()
		}

		for {
			select {
			case <-abortCh:
				log.Println("INFO clientFactory() received abort signal, quiting...")

				for _, client := range clients {
					close(client.abortCh)
				}

				// wait for all active clients to exit
				wg.Wait()
				close(doneCh)
				return
			case <-time.After(time.Millisecond * 500):
				continue
			}
		}
	}(doneCh, abortCh, output)
	return doneCh
}

func collectResults(abort chan bool, input chan *webResponse) chan bool {
	doneCh := make(chan bool)
	go func(doneCh chan bool, abortCh chan bool, inputCh chan *webResponse) {
		res := &results{}
		statsTicker := time.Tick(time.Second * 1)

		// Create a new HTTPClient for influxdb
		c, _ := influxc.NewHTTPClient(influxc.HTTPConfig{
			Addr: "http://localhost:8086",
		})
		// Create a new point batch
		bp, err := influxc.NewBatchPoints(influxc.BatchPointsConfig{
			Database:  InfluxDbName,
			Precision: "ns",
		})
		if err != nil {
			log.Println("collectResults() failed to setup BatchPoints", err)
		}

		for {
			select {
			case msg := <-input:
				res.add(msg)
				isSuccess := "false"
				reqDurationMs := 100.0

				if msg.Error == nil {
					isSuccess = "true"
					dur, _ := time.ParseDuration(msg.ReqTime)
					reqDurationMs = float64(dur.Nanoseconds() / time.Millisecond.Nanoseconds())
				}

				tags := map[string]string{"app-version": msg.AppVersion, "success": isSuccess}
				fields := map[string]interface{}{
					"responseTime":  reqDurationMs,
					"totalRequests": 1.0,
				}
				pt, _ := influxc.NewPoint("netRequests", tags, fields, time.Now())
				bp.AddPoint(pt)
			case <-abortCh:
				log.Println("INFO collectResults() received abort signal, quiting...")
				close(doneCh)
				return
			case <-statsTicker:
				log.Printf("INFO collectResults() STATS: %s\n", res)
				res.clear()

				// Write the batch
				if err := c.Write(bp); err != nil {
					log.Println("collectResults() failed to write batch", err)
				}
				bp, _ = influxc.NewBatchPoints(influxc.BatchPointsConfig{
					Database:  InfluxDbName,
					Precision: "ns",
				})
			}
		}
	}(doneCh, abort, input)
	return doneCh
}

func main() {
	log.Println("INFO horde starting up with max connections set to", maxConn)
	log.Println("INFO target host set to", remoteHostPort)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	collectorAbortCh := make(chan bool, 1)
	clientAbortCh := make(chan bool, 1)
	dataCh := make(chan *webResponse)

	// start collecting results
	collectorDone := collectResults(collectorAbortCh, dataCh)

	// start sending requests
	clientDone := clientFactory(clientAbortCh, dataCh)

	// wait for interrupt signal
	<-signalCh

	// notify client factory to stop sending requests
	close(clientAbortCh)

	// wait for client factory to exit cleanly
	<-clientDone
	log.Println("INFO: clientFactory() exited")

	// notify collector to stop collecting results
	close(collectorAbortCh)

	// wait for collector to exit cleanly
	<-collectorDone
	log.Println("INFO: collectResults() exited")

	// be fancy and close all resources
	close(dataCh)

	log.Println("INFO: application exit")
}
