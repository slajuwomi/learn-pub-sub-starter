package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	fmt.Println("Starting Peril server...")
	conString := "amqp://guest:guest@localhost:5672/"
	con, err := amqp.Dial(conString)
	if err != nil {
		fmt.Println("Server dial failed: ", err)
	}
	defer con.Close()
	fmt.Println("Connection successful")

	channel, err := con.Channel()
	if err != nil {
		fmt.Println("Channel creation failed", err)
	}
	// FAKE

	_, _, err = pubsub.DeclareAndBind(con, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", "durable")
	if err != nil {
		fmt.Println("Failed to declare and bind queue: ", err)
	}

	gamelogic.PrintServerHelp()
	for {
		userInput := gamelogic.GetInput()
		if userInput == nil {
			continue
		}

		firstWord := userInput[0]
		switch firstWord {
		case "pause":
			fmt.Println("Sending a pause message")
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		case "resume":
			fmt.Println("Sending a resume message")
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		case "quit":
			fmt.Println("Exiting")
			return
		default:
			fmt.Println("Invalid command")
		}
	}

}
