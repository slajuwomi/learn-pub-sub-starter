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
			message := fmt.Sprintf("%v won a war against %v\n", winner, loser)
			fmt.Println(message)
			thisAcktype = PublishGameLog(publishCh, message, rw.Attacker.Username)
		case gamelogic.WarOutcomeDraw:
			message := fmt.Sprintf("A war between %v and %v resulted in a draw\n", winner, loser)
			fmt.Println(message)
			thisAcktype = PublishGameLog(publishCh, message, rw.Attacker.Username)
		case gamelogic.WarOutcomeYouWon:
			fmt.Printf("%v won a war against %v\n", winner, loser)
		default:
			fmt.Println("Failed to get vaild war outcome")
			thisAcktype = pubsub.NackDiscard
		}
		return thisAcktype
	}
}

func PublishGameLog(ch *amqp091.Channel, message, username string) pubsub.Acktype {
	var newGameLog routing.GameLog
	newGameLog.CurrentTime = time.Now()
	newGameLog.Message = message
	newGameLog.Username = username
	err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+username, newGameLog)
	if err != nil {
		return pubsub.NackRequeue
	}
	return pubsub.Ack
}
