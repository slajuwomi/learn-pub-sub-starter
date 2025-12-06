package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerWarMessages(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	var thisAcktype pubsub.Acktype
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(rw)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			thisAcktype = pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			thisAcktype = pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			// message := fmt.Sprintf("%v won a war against %v\n", winner, loser)
			// fmt.Println(message)
			err := PublishGameLog(publishCh, fmt.Sprintf("%v won a war against %v\n", winner, loser), gs.GetUsername())
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.NackRequeue
			}
			// thisAcktype = PublishGameLog(publishCh, message, rw.Attacker.Username)
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			// message := fmt.Sprintf("A war between %v and %v resulted in a draw\n", winner, loser)
			// fmt.Println(message)
			err := PublishGameLog(publishCh, fmt.Sprintf("A war between %v and %v resulted in a draw\n", winner, loser), gs.GetUsername())
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.NackRequeue
			}
			// thisAcktype = PublishGameLog(publishCh, message, rw.Attacker.Username)
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			err := PublishGameLog(publishCh, fmt.Sprintf("%v won a war against %v", winner, loser), gs.GetUsername())
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("Failed to get vaild war outcome")
			thisAcktype = pubsub.NackDiscard
		}
		fmt.Println("error: unkonwn war outcome")
		return thisAcktype
	}
}

func PublishGameLog(ch *amqp091.Channel, message, username string) error {
	return pubsub.PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     message,
		},
	)
}
