package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to create channel: %v", err)
	}

	queue, err := channel.QueueDeclare(queueName, queueType == Durable, queueType == Transient, queueType == "transient", false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to declare queue: %v", err)
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to bind queue to exchange: %v", err)
	}
	return channel, queue, nil

}
