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

// Package messaging contains middleware for communication with decoupled
// services, could be RabbitMQ drivers as well as kafka or redis
package messaging

import (
	"errors"

	"github.com/streadway/amqp"
)

// ErrRabbitMq generic RabbitMQ communication error
var ErrRabbitMq = errors.New("rabbitmq communication error")

// MessageQueue defines the behavior of a simple message queue, it's
// expected to provide a `Produce` function a `Consume` one and a `Close`.
type MessageQueue interface {
	Produce(string, []byte) error
	Consume(string, int, chan<- []byte) error
	Close()
}

// AmqpOptions is a simple settings container for AMQP queue
type amqpOptions struct {
	durable      bool
	deleteUnused bool
	exclusive    bool
	noWait       bool
}

// amqpOption is an option pattern helper to set different options to an
// `amqpOption` object
type amqpOption func(*amqpOptions)

// Connect create a connection and a channel for RabbitMQ communication,
// returning them into a pointer to an `AmqpQueue` object
func Connect(url string, opts ...amqpOption) (*AmqpQueue, error) {
	options := &amqpOptions{}

	// Mix in all optionals
	for _, opt := range opts {
		opt(options)
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &AmqpQueue{conn, options, channel}, nil
}

// AmqpQueue is the main exposed object to work with, it's a `MessageQueue`
// object
type AmqpQueue struct {
	amqpConn   *amqp.Connection
	connection *amqpOptions
	channel    *amqp.Channel
}

// Close a connection with RabbitMQ by closing the underlying connection and
// the channel
func (q AmqpQueue) Close() {
	q.channel.Close()
	q.amqpConn.Close()
}

// Produce publish a message to a define queue name
func (q AmqpQueue) Produce(queueName string, item []byte) error {
	queue, err := q.channel.QueueDeclare(
		queueName,                 // name
		q.connection.durable,      // durable
		q.connection.deleteUnused, // delete when unused
		q.connection.exclusive,    // exclusive
		q.connection.noWait,       // no-wait
		nil,                       // arguments
	)
	if err != nil {
		return err
	}

	err = q.channel.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        item,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// Consume subscribe to a queue and block consuming all messages incoming, a
// concurrency value can be set to consume multiple messages at once
func (q AmqpQueue) Consume(queueName string, concurrency int,
	itemChan chan<- []byte) error {
	queue, err := q.channel.QueueDeclare(
		queueName,                 // name
		q.connection.durable,      // durable
		q.connection.deleteUnused, // delete when unused
		q.connection.exclusive,    // exclusive
		q.connection.noWait,       // no-wait
		nil,                       // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := q.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return err
	}

	// pre-fetch `concurrency` message at once
	err = q.channel.Qos(concurrency, 0, false)
	if err != nil {
		return err
	}

	forever := make(chan error)

	// Forever consume messages and push them into the channel for the client
	go func() {
		for d := range msgs {
			itemChan <- d.Body
		}
		forever <- ErrRabbitMq
	}()

	err = <-forever
	return err
}
