package main

import (
	"fmt"
	"github.com/bennytemmerman/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bennytemmerman/learn-pub-sub-starter/internal/pubsub"
	"github.com/bennytemmerman/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func main() {
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	// Declare the peril_topic exchange (topic exchange)
	err = publishCh.ExchangeDeclare(
		routing.ExchangePerilTopic, // name of the exchange
		"topic",                    // type
		true,                       // durable
		false,                      // auto-deleted
		false,                      // internal
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		log.Fatalf("could not declare exchange: %v", err)
	}

	// Declare the durable queue game_logs
	queue, err := publishCh.QueueDeclare(
		"game_logs", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Fatalf("could not declare queue: %v", err)
	}

	// Bind the queue to the topic exchange with routing key game_logs.*
	err = publishCh.QueueBind(
		queue.Name,                 // name of the queue
		routing.GameLogSlug+".*",  // routing key (e.g., game_logs.*)
		routing.ExchangePerilTopic, // exchange
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		log.Fatalf("could not bind queue: %v", err)
	}

	fmt.Println("Queue 'game_logs' bound to exchange 'peril_topic' with routing key 'game_logs.*'")


	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("Publishing paused game state")
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "resume":
			fmt.Println("Publishing resumes game state")
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "quit":
			log.Println("goodbye")
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
