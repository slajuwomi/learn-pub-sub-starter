package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackDiscard
	NackRequeue
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

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) Acktype) error {
	chann, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange: %v", err)
	}
	err = chann.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("failed to qos: %v", err)
	}
	deliveryChannel, err := chann.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consumption: %v", err)
	}

	go func() error {
		defer chann.Close()
		for delivery := range deliveryChannel {
			var unmarshaledDelivery T
			err := json.Unmarshal(delivery.Body, &unmarshaledDelivery)
			if err != nil {
				return err
			}
			switch handler(unmarshaledDelivery) {
			case Ack:
				err = delivery.Ack(false)
				if err != nil {
					return err
				}
				fmt.Println("Ack occurred")
			case NackDiscard:
				err = delivery.Nack(false, false)
				if err != nil {
					return err
				}
				fmt.Println("NackDiscard occurred")
			case NackRequeue:
				err = delivery.Nack(false, true)
				if err != nil {
					return err
				}
				fmt.Println("NackRequeue occurred")
			}
		}
		return nil
	}()
	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(val)
	if err != nil {
		return fmt.Errorf("failed to encode to gob: %v", err)
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        buffer.Bytes(),
	})
}

func SubscribeGob[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) Acktype) error {
	chann, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange: %v", err)
	}
	err = chann.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("failed to qos: %v", err)
	}
	deliveryChannel, err := chann.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consumption: %v", err)
	}

	go func() error {
		defer chann.Close()
		for delivery := range deliveryChannel {
			var decoded T
			buffer := bytes.NewBuffer(delivery.Body)
			decoder := gob.NewDecoder(buffer)

			err = decoder.Decode(&decoded)
			if err != nil {
				fmt.Printf("Error decoding message: %v/n", err)
				delivery.Nack(false, false)
				continue
			}
			switch handler(decoded) {
			case Ack:
				err = delivery.Ack(false)
				if err != nil {
					return err
				}
				fmt.Println("Ack occurred")
			case NackDiscard:
				err = delivery.Nack(false, false)
				if err != nil {
					return err
				}
				fmt.Println("NackDiscard occurred")
			case NackRequeue:
				err = delivery.Nack(false, true)
				if err != nil {
					return err
				}
				fmt.Println("NackRequeue occurred")
			}
		}
		return nil
	}()
	return nil
}
