package main

import (
	"fmt"
	"net/http"
	"time"
)

type ServerStats struct {
	Url                string
	Alive              bool
	LatestResponseTime time.Duration
	ResponseStatus     string
}

func probeServer(serverChan <-chan string, statsChan chan<- *ServerStats, stopChan <-chan interface{}) {
	for {
		select {
		case server := <-serverChan:
			stats := &ServerStats{server, false, 0, ""}
			start := time.Now()
			res, err := http.Get(server)
			elapsed := time.Since(start)
			if err == nil {
				stats.Alive = true
				stats.ResponseStatus = res.Status
			}
			stats.LatestResponseTime = elapsed
			statsChan <- stats
		case <-stopChan:
			break
		}
	}
}

func main() {
	servers := []string{"http://localhost:7892", "http://localhost:9898"}
	serverChannel := make(chan string)
	statsChannel := make(chan *ServerStats)
	stopChannel := make(chan interface{})

	go func() {
		for {
			stats := <-statsChannel
			fmt.Printf("%v\n", stats)
		}
	}()

	for range servers {
		go probeServer(serverChannel, statsChannel, stopChannel)
	}

	for {
		for _, server := range servers {
			serverChannel <- server
		}
		time.Sleep(5000 * time.Millisecond)
	}

}
