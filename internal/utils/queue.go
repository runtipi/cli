package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EventData struct {
	QueueName   string `json:"queueName"`
	RequestData any    `json:"requestData"`
	RPCStatus   string `json:"rpcStatus"`
	RPCMessage  string `json:"rpcMessage"`
	Timestamp   string `json:"timestamp"`
	WorkerID    string `json:"workerId,omitempty"`
	RequestID   string `json:"requestId,omitempty"`
}

type QueueOptions struct {
	ConnectionURL string
	Exchange      string
	RoutingKey    string
	Filter        func(event EventData) bool
	Timeout       time.Duration
}

func NewDefaultQueueOptions() QueueOptions {
	password := GetEnvValue("RABBITMQ_PASSWORD")

	return QueueOptions{
		ConnectionURL: fmt.Sprintf("amqp://tipi:%s@localhost:5672/", password),
		Exchange:      "app-events-queue",
		RoutingKey:    "rpc.#",
	}
}

func ListenForMessage(ctx context.Context, options QueueOptions, resultChan chan<- EventData) {
	conn, err := amqp.Dial(options.ConnectionURL)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()
		conn.Close()
		ch.Close()
	}()

	var timeoutChan <-chan time.Time
	if options.Timeout > 0 {
		timeoutChan = time.After(options.Timeout)
	}

	err = ch.ExchangeDeclare(
		options.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange '%s': %v", options.Exchange, err)
	}

	observerQueueName := fmt.Sprintf("go_app_events_observer_%d", time.Now().UnixNano())
	q, err := ch.QueueDeclare(
		observerQueueName,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue '%s': %v", observerQueueName, err)
	}

	err = ch.QueueBind(
		q.Name,
		options.RoutingKey,
		options.Exchange,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue '%s' to exchange '%s' with routing key '%s': %v",
			q.Name, options.Exchange, options.RoutingKey, err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer on queue '%s': %v", q.Name, err)
	}

	go func() {
		defer func() {
			ch.Close()
			conn.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case <-timeoutChan:
				return

			case d, ok := <-msgs:
				if !ok {
					return
				}

				var event EventData
				err := json.Unmarshal(d.Body, &event)
				if err != nil {
					if nackErr := d.Nack(false, false); nackErr != nil {
						log.Printf("Failed to nack message after JSON unmarshal error: %v", nackErr)
					}
					continue
				}

				if options.Filter != nil && !options.Filter(event) {
					if ackErr := d.Ack(false); ackErr != nil {
						log.Printf("Failed to ack filtered message: %v", ackErr)
					}
					continue
				}

				if ackErr := d.Ack(false); ackErr != nil {
					log.Printf("Failed to ack processed message: %v", ackErr)
				}
				select {
				case resultChan <- event:
				case <-ctx.Done():
				}
				return
			}
		}
	}()
}

func WaitForEvent(ctx context.Context, timeout time.Duration, filter func(EventData) bool) (*EventData, bool) {
	options := NewDefaultQueueOptions()
	options.Timeout = timeout
	options.Filter = filter

	resultChan := make(chan EventData, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go ListenForMessage(ctx, options, resultChan)

	select {
	case event := <-resultChan:
		return &event, event.RPCStatus == "success"
	case <-time.After(timeout):
		return nil, false
	case <-ctx.Done():
		return nil, false
	}
}
