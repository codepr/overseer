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
// aggregations and analysis of incoming server statistics and middleware to
// forward it to front-end clients
package backend

import (
	"log"
	"net/http"

	"github.com/codepr/overseer/internal/messaging"

	"github.com/gorilla/websocket"
)

// Gorilla websocket upgrader for HTTP endpoints
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// wsEndpoint is an HTTP route that converts connected clients to websocket
// connections, streaming record stats coming from RabbitMQ
func wsEndpoint(events <-chan []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
		}

		// Push events from channel directly to the connected client
		for {
			event := <-events
			if err := ws.WriteMessage(websocket.TextMessage, event); err != nil {
				log.Println(err)
				return
			}

		}
	}
}

// Run start consuming the RabbitMQ and run the HTTP server serving `ws_stats`
// as the only route available
func Run(listenAddr, queueAddr, queueName string) {
	events := make(chan []byte)
	queue, err := messaging.Connect(queueAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer queue.Close()

	// Add websocket route
	http.HandleFunc("/ws_stats", wsEndpoint(events))

	// Consume records from RabbitMQ pushing them to `events` channel
	go func() {
		queue.Consume(queueName, 1, events)
	}()

	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
