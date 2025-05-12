package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// QueueType enum
const (
	TransientQueue = iota
	DurableQueue
)

// SubscribeJSON subscribes to a queue, consuming and decoding JSON messages of type T.
func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T),
) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("could not open channel: %w", err)
	}

	err = DeclareAndBind(ch, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %w", err)
	}

	deliveries, err := ch.Consume(
		queueName,
		"",    // auto-generated consumer name
		false, // auto-ack (we want to manually ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("could not start consuming: %w", err)
	}

	go func() {
		for msg := range deliveries {
			var t T
			if err := json.Unmarshal(msg.Body, &t); err != nil {
				log.Printf("could not unmarshal message: %v", err)
				continue
			}
			handler(t)
			msg.Ack(false)
		}
	}()

	return nil
}
