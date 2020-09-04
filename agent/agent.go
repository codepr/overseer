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

// Package agent defines and implement the probing URLs servers for health
// utilities
package agent

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"

	. "github.com/codepr/overseer/internal"
)

// Agent is responsible for probing a list of URLs once every `interval`
// milliseconds and collecting some statistics like response time, status code
// and content of the queried server.
//
// Finally it forwards every result of the call to a middleware, generally a
// message queue that can be consumed by other services.
type Agent struct {
	urls     []URL
	interval time.Duration
	timeout  time.Duration
	mq       ProducerConsumer
	logger   *log.Logger
}

// conf is a private configuration object, just act as a container for user
// defined settings read from yaml file on the filesystem
type conf struct {
	Agent struct {
		Servers    []URL         `yaml:"servers"`
		Interval   time.Duration `yaml:"interval"`
		Timeout    time.Duration `yaml:"timeout"`
		WindowSize int           `yaml:"window_size"`
		AmqpAddr   string        `yaml:"amqp_addr,omitempty"`
		QueueName  string        `yaml:"queue_name,omitempty"`
	} `yaml:"agent"`
}

// loadConf initialize and return a pointer to `conf` struct, by reading
// key-values from a yaml file on the filesystem
func loadConf(path string) (*conf, error) {
	// Set some default values
	config := new(conf)
	config.Agent.AmqpAddr = "amqp://guest:guest@localhost:5672/"
	config.Agent.QueueName = "urlstatus"
	// Override with FS configuration read
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// NewFromConfig create a new `Agent` and return a pointer to it by loading
// the configuration from the filesystem through `loadConf` call
func NewFromConfig(path string) (*Agent, error) {
	conf, err := loadConf(path)
	if err != nil {
		return nil, err
	}
	// Create a new message queue
	mq := NewAmqpQueue(conf.Agent.AmqpAddr, conf.Agent.QueueName)
	agent := New(conf.Agent.Servers, conf.Agent.WindowSize,
		conf.Agent.Interval, conf.Agent.Timeout, mq)
	return agent, nil
}

// New create a new `Agent` and return a pointer to it
func New(urls []URL, windowSize int, interval,
	timeout time.Duration, mq ProducerConsumer) *Agent {
	return &Agent{
		urls:     urls,
		interval: interval,
		timeout:  timeout,
		mq:       mq,
		logger:   log.New(os.Stdout, "agent: ", log.LstdFlags),
	}
}

// Run start the main `Agent` process, a loop probing registered URLs and
// sending results to a message queue
func (a *Agent) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	urlChan := make(chan URL)

	a.logger.Println("Monitoring agent starting")
	a.logger.Printf("Refresh interval: %v\n", a.interval)
	a.logger.Printf("Request timeout: %v\n", a.timeout)
	a.logger.Println("Monitoring servers:")

	for _, url := range a.urls {
		a.logger.Printf("  - %v\n", url)
	}

	// Spawn a goroutine for each monitored URL that will probe it once every
	// `interval` seconds
	for range a.urls {
		go func(ctx context.Context) {
			for {
				select {
				case url := <-urlChan:
					status, _ := probeServer(url)
					// Encode the status retrieved from the HTTP healthcheck
					// call and send it into the AMQP queue to the aggregator
					// service
					payload, err := json.Marshal(status)
					if err != nil {
						a.logger.Println("Error encoding status")
						continue
					}
					if err := a.mq.Produce(payload); err != nil {
						a.logger.Println("Error producing status to queue")
					}
				case <-ctx.Done():
					return
				}
			}
		}(ctx)
	}

	// Graceful shutdown of workers
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		cancel()
		os.Exit(1)
	}()

	// Loop through URLs and send them to a worker goroutine to probe them,
	// once every `interval` milliseconds
	for {
		for _, url := range a.urls {
			urlChan <- url
		}
		time.Sleep(a.interval)
	}
}

// probeServer perform an HTTP GET request to an URL, tracking response time
// status code and content
func probeServer(url URL) (*ServerStatus, error) {
	status := &ServerStatus{Url: url, Alive: true}
	// Clock the response time
	start := time.Now()
	res, err := http.Get(url)
	elapsed := time.Since(start)
	// If something goes wrong with the HTTP call just set the server
	// as offline with 500 err
	if err != nil {
		status.Alive = false
		status.ResponseStatus = http.StatusInternalServerError
	} else {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		// If **NO errors** happens reading the Body content, set the
		// status Body to the content read
		if err == nil {
			status.ResponseContent = string(body)
		}
		status.ResponseStatus = res.StatusCode
	}
	status.ResponseTime = elapsed
	return status, nil
}
