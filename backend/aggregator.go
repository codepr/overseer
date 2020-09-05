// BSD 2-Clause License
//
// Copyright (c) 2020, Andrea Giacomo Baldan
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package backend contains all backend related modules and utilies to perform
// aggregations and analysis of incoming server statistics
package backend

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	. "github.com/codepr/overseer/internal"
	"github.com/codepr/overseer/internal/messaging"
)

// serverStats act as simple ephemeral state for each probed URL server,
// tracking its status, availability and some stats like the moving average
// over last N points
type serverStats struct {
	Alive              bool
	MovingAverageStats *MovingAverage
	LatestResponseTime time.Duration
	ResponseStatusMap  map[int]int
	Availability       float64
}

// Aggregator performs some aggregation on incoming records from a message queue
// tracking the states on a map
type Aggregator struct {
	servers    map[URL]*serverStats
	windowSize int
	mq         messaging.MessageQueue
	logger     *log.Logger
}

// NewAggregator create a new `Aggregator` object
func NewAggregator() *Aggregator {
	mq, _ := messaging.Connect("amqp://guest:guest@localhost:5672/")
	return &Aggregator{
		servers:    make(map[URL]*serverStats),
		windowSize: 120,
		mq:         mq,
		logger:     log.New(os.Stdout, "aggregator: ", log.LstdFlags),
	}
}

// Run start the consume process from the message queue and aggregation of
// incoming records
func (a *Aggregator) Run() {
	events := make(chan []byte)
	urls := make(chan URL)
	ctx, cancel := context.WithCancel(context.Background())

	// Catch SIGINT/SIGTERM signals and call cancel() before exiting to
	// gracefully stop goroutines
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalCh
		cancel()
		os.Exit(1)
	}()

	// Run an event listener goroutine, compute aggregation on `ServerStatus`
	// events coming from the message queue
	go func(ctx context.Context) {
		for {
			select {
			case event := <-events:
				var status ServerStatus
				err := json.Unmarshal(event, &status)
				if err != nil {
					a.logger.Println("Error decoding status event")
				} else {
					a.aggregate(&status)
					urls <- status.Url
				}
			case <-ctx.Done():
				a.mq.Close()
				return
			}
		}
	}(ctx)

	// Just print results of aggreation fo each received URL
	go func(ctx context.Context) {
		for {
			select {
			case url := <-urls:
				stats, _ := a.servers[url]
				a.logger.Printf("%s alive=%v avail.(%%)=%.2f res(ms)=%v min(ms)=%v max(ms)=%v avg(ms)=%v status_codes=%v\n",
					url, stats.Alive, stats.Availability,
					stats.LatestResponseTime, stats.MovingAverageStats.Min(),
					stats.MovingAverageStats.Max(), stats.MovingAverageStats.Mean(),
					stats.ResponseStatusMap)
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	if err := a.mq.Consume("urlstatus", 1, events); err != nil {
		a.logger.Fatal(err)
	}
}

// Perform some aggregations by adding new received records to previous history
// for a given URL
func (a *Aggregator) aggregate(status *ServerStatus) {
	// Check if the URL is already mapped, add a new `ServerStats` pointer
	// if empty
	if stats, ok := a.servers[status.Url]; !ok {
		a.servers[status.Url] = &serverStats{
			false,
			NewMovingAverage(a.windowSize),
			0,
			map[int]int{},
			0.0,
		}
	} else {
		stats.Alive = status.Alive
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
