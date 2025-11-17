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
