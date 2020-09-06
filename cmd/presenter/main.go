package main

import (
	"flag"
	"github.com/codepr/overseer/backend"
)

var (
	listenAddr string
	queueAddr  string
	queueName  string
)

func main() {
	flag.StringVar(&listenAddr, "listen", ":17657", "Addr to listen on for websocket exposed API")
	flag.StringVar(&queueAddr, "queue_addr", "amqp://guest:guest@localhost:5672", "Addr of the RabbitMQ middleware to communicate with agreagator")
	flag.StringVar(&queueName, "queue", "stats", "Name of the queue to consume")
	flag.Parse()
	backend.Run(listenAddr, queueAddr, queueName)
}
