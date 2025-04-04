// a package for working with rabbit mq
package mb

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageBroker struct
type MessageBroker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

// close message broker connection
func (message_broker *MessageBroker) Close() {
	message_broker.ch.Close()
	message_broker.conn.Close()
	log.Printf("closed message broker connection\n")
}

// InitMessageBroker initializes a new message broker connection
func InitMessageBroker(addr, user, pass, q_name string) (*MessageBroker, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s", user, pass, addr)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		q_name,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &MessageBroker{
		conn, ch, q,
	}, nil
}

// ProduceTextMsg sends a text message to the queue
func ProduceTextMsg(mb *MessageBroker, msg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mb.ch.PublishWithContext(
		ctx,
		"",
		mb.q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	log.Printf("sent msg: %s\n", msg)
	return nil
}

// runs a consumer that listens for messages on the queue
func RunConsumer(mb *MessageBroker, handler func(data []byte)) error {
	msgs, err := mb.ch.Consume(
		mb.q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		for d := range msgs {
			log.Printf("reading message!\n")
			handler(d.Body)
		}
	}()

	log.Printf("waiting for message...\n")
	<-ctx.Done()
	log.Printf("shuting down consumer...\n")
	return nil
}
