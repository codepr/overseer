package agent

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	. "github.com/codepr/overseer/internal"
)

type Agent struct {
	urls     []URL
	interval time.Duration
	timeout  time.Duration
	mq       ProducerConsumer
	logger   *log.Logger
}

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

func (a *Agent) Run() {
	urlChan := make(chan URL)
	stopChan := make(chan interface{})

	log.Println("Monitoring agent starting")
	log.Printf("Refresh interval: %v\n", a.interval)
	log.Printf("Request timeout: %v\n", a.timeout)
	log.Println("Monitoring servers:")
	for _, url := range a.urls {
		log.Printf("  - %v\n", url)
	}

	for range a.urls {
		go func() {
			for {
				select {
				case url := <-urlChan:
					status, _ := probeServer(url)
					// Encode the status retrieved from the HTTP healthcheck call and send
					// it into the AMQP queue to the aggregator service
					payload, err := json.Marshal(status)
					if err != nil {
						a.logger.Println("Error encoding status")
						continue
					}
					if err := a.mq.Produce(payload); err != nil {
						a.logger.Println("Error producing status to queue")
					}
				case <-stopChan:
					break
				}
			}
		}()
	}

	for {
		for _, url := range a.urls {
			urlChan <- url
		}
		time.Sleep(a.interval)
	}
}

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
