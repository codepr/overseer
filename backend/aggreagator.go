package backend

import (
	"encoding/json"
	"log"
	"os"
	"time"

	. "github.com/codepr/overseer/internal"
)

type ServerStats struct {
	Alive              bool
	MovingAverageStats *MovingAverage
	LatestResponseTime time.Duration
	ResponseStatusMap  map[int]int
	Availability       float64
}

type Aggregator struct {
	servers    map[URL]*ServerStats
	windowSize int
	mq         ProducerConsumer
	logger     *log.Logger
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		servers:    make(map[URL]*ServerStats),
		windowSize: 120,
		mq:         NewAmqpQueue("amqp://guest:guest@localhost:5672/", "urlstatus"),
		logger:     log.New(os.Stdout, "aggregator: ", log.LstdFlags),
	}
}

func (a *Aggregator) Run() error {
	events := make(chan []byte)
	urls := make(chan URL)

	// Run an event listener goroutine, compute aggregation on `ServerStatus`
	// events coming from the message queue
	go func() {
		for {
			event := <-events
			var status ServerStatus
			err := json.Unmarshal(event, &status)
			if err != nil {
				a.logger.Println("Error decoding status event")
			} else {
				a.aggregate(&status)
				urls <- status.Url
			}
		}
	}()

	// Just print results of aggreation fo each received URL
	go func() {
		for {
			url := <-urls
			stats, _ := a.servers[url]
			a.logger.Printf("%s alive=%v avail.(%%)=%.2f res(ms)=%v min(ms)=%v max(ms)=%v avg(ms)=%v status_codes=%v\n",
				url, stats.Alive, stats.Availability,
				stats.LatestResponseTime, stats.MovingAverageStats.Min(),
				stats.MovingAverageStats.Max(), stats.MovingAverageStats.Mean(),
				stats.ResponseStatusMap)
		}
	}()

	return a.mq.Consume(events)
}

func (a *Aggregator) aggregate(status *ServerStatus) {
	// Check if the URL is already mapped, add a new `ServerStats` pointer
	// if empty
	if stats, ok := a.servers[status.Url]; !ok {
		a.servers[status.Url] = &ServerStats{
			false,
			NewMovingAverage(a.windowSize),
			0,
			map[int]int{},
			0.0,
		}
	} else {
		stats.ResponseStatusMap[status.ResponseStatus] += 1
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
		stats.LatestResponseTime = status.ResponseTime
		stats.MovingAverageStats.Put(status.ResponseTime)
		a.servers[status.Url] = stats
	}
}
