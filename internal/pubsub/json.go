package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ah *amqp.Channel, exchange, key string, val T) error {
	valJsonBytes, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal val: %v", err)
	}

	ah.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        valJsonBytes,
	})
	return nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T)) error {
	chann, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange: %v", err)
	}
	deliveryChannel, err := chann.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consumption: %v", err)
	}

	go func() error {
		for delivery := range deliveryChannel {
			var unmarshaledDelivery T
			err := json.Unmarshal(delivery.Body, &unmarshaledDelivery)
			if err != nil {
				return err
			}
			handler(unmarshaledDelivery)
			err = delivery.Ack(false)
			if err != nil {
				return err
			}

		}
		return nil
	}()
	return nil
}
