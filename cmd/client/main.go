package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	conString := "amqp://guest:guest@localhost:5672/"
	con, err := amqp.Dial(conString)
	if err != nil {
		fmt.Println("Client dial failed: ", err)
	}
	defer con.Close()
	fmt.Println("Connection successful")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Failed to get username: ", err)
	}

	chann, _, err := pubsub.DeclareAndBind(con, routing.ExchangePerilDirect, fmt.Sprintf(routing.PauseKey+"."+username), routing.PauseKey, "transient")
	if err != nil {
		fmt.Println("Failed to declare and bind queue: ", err)
	}

	gameState := gamelogic.NewGameState(username)

	_, err = pubsub.SubscribeJSON(con, routing.ExchangePerilDirect, fmt.Sprintf(routing.PauseKey+"."+username), routing.PauseKey, "transient", handlerPause(gameState))
	if err != nil {
		fmt.Println("Failed to subscribe: ", err)
	}

	_, err = pubsub.SubscribeJSON(con, routing.ExchangePerilTopic, fmt.Sprintf(routing.ArmyMovesPrefix+"."+username), fmt.Sprintf(routing.ArmyMovesPrefix+"."+"*"), "transient", handlerMove(gameState))
	if err != nil {
		fmt.Println("Failed to subscribe: ", err)
	}

	for {
		userInput := gamelogic.GetInput()
		if userInput == nil {
			continue
		}

		firstWord := userInput[0]
		switch firstWord {
		case "spawn":
			err = gameState.CommandSpawn(userInput)
			if err != nil {
				fmt.Println("Spawning failed: ", err)
			}
		case "move":
			move, err := gameState.CommandMove(userInput)
			if err != nil {
				fmt.Println("Failed to move: ", err)
			} else {
				err = pubsub.PublishJSON(chann, routing.ExchangePerilTopic, fmt.Sprintf(routing.ArmyMovesPrefix+"."+username), move)
				if err != nil {
					fmt.Println("Move unsuccessful! ", err)
				} else {
					fmt.Println("Move successful!")
				}
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Invalid command")
		}
	}
}
