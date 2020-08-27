package overseer

import (
	"fmt"
	"net/http"
	"time"
)

type ServerStats struct {
	Url                string
	Alive              bool
	MovingAverageStats *MovingAverage
	LatestResponseTime time.Duration
	ResponseStatusMap  map[int]int
	Availability       float64
}

type Agent struct {
	servers  []*ServerStats
	interval time.Duration
}

func NewAgent(urls []string, windowSize int, interval time.Duration) *Agent {
	servers := make([]*ServerStats, len(urls))
	for i, url := range urls {
		servers[i] = &ServerStats{
			url,
			false,
			NewMovingAverage(windowSize),
			0,
			map[int]int{},
			0.0,
		}

	}
	return &Agent{
		servers:  servers,
		interval: interval * time.Millisecond,
	}
}

func (a *Agent) Run() {
	serverChannel := make(chan *ServerStats)
	statsChannel := make(chan *ServerStats)
	stopChannel := make(chan interface{})

	go func() {
		for {
			stats := <-statsChannel
			fmt.Printf("%s alive=%v avail.(%%)=%.2f res(ms)=%v min(ms)=%v max(ms)=%v avg(ms)=%v status_codes=%v\n",
				stats.Url, stats.Alive, stats.Availability,
				stats.LatestResponseTime, stats.MovingAverageStats.Min(),
				stats.MovingAverageStats.Max(), stats.MovingAverageStats.Mean(),
				stats.ResponseStatusMap)
		}
	}()

	for range a.servers {
		go probeServer(serverChannel, statsChannel, stopChannel)
	}

	for {
		for _, server := range a.servers {
			serverChannel <- server
		}
		time.Sleep(5000 * time.Millisecond)
	}
}

func probeServer(serverChan <-chan *ServerStats, statsChan chan<- *ServerStats, stopChan <-chan interface{}) {
	for {
		select {
		case stats := <-serverChan:
			start := time.Now()
			res, err := http.Get(stats.Url)
			elapsed := time.Since(start)
			if err == nil {
				stats.Alive = true
				stats.ResponseStatusMap[res.StatusCode] += 1
			} else {
				stats.Alive = false
				stats.ResponseStatusMap[500] += 1
			}
			// Retrieve availability ratio % by counting error codes and
			// valid codes
			validCodes, errorCodes := 0, 0
			for k, v := range stats.ResponseStatusMap {
				if k > 400 {
					errorCodes += v
				} else {
					validCodes += v
				}
			}
			stats.Availability = float64(validCodes*100.0) / float64((validCodes + errorCodes))
			stats.LatestResponseTime = elapsed
			stats.MovingAverageStats.Put(elapsed)
			statsChan <- stats
		case <-stopChan:
			break
		}
	}
}
