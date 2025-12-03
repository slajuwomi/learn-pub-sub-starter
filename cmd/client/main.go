package main

import (
	"fmt"
	"strconv"

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

	err = pubsub.SubscribeJSON(con, routing.ExchangePerilDirect, fmt.Sprintf(routing.PauseKey+"."+username), routing.PauseKey, "transient", handlerPause(gameState))
	if err != nil {
		fmt.Println("Failed to subscribe: ", err)
	}

	err = pubsub.SubscribeJSON(con, routing.ExchangePerilTopic, fmt.Sprintf(routing.ArmyMovesPrefix+"."+username), fmt.Sprintf(routing.ArmyMovesPrefix+"."+"*"), "transient", handlerMove(gameState, chann, username))

	err = pubsub.SubscribeJSON(con, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, fmt.Sprintf(routing.WarRecognitionsPrefix+"."+"*"), "durable", handlerWarMessages(gameState, chann))

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
			if userInput[1] != "" {
				numberOfSpams, err := strconv.Atoi(userInput[1])
				if err != nil {
					fmt.Println("Failed to convert spam number: ", err)	
					return
				}
				for i :=0; i < numberOfSpams; i++ {
					maliciousLogMessage := gamelogic.GetMaliciousLog()
					err := pubsub.PublishJSON(chann, routing.ExchangePerilTopic, routing.GameLogSlug+"."+username, maliciousLogMessage)
					if err != nil {
						fmt.Println("Failed to publish spam to game logs: ", err)
						return
					}
				}
			}
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Invalid command")
		}
	}
}
