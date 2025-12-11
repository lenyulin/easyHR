package aiagentmanager

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AgentMsgProducer interface {
	AgentProduceResponseReceivedEvent(evt ResponseReceivedEvent) error
}

type ResponseReceivedEvent struct {
	SessionID string
	ModelType string
	Role      string
	Input     string
	Output    string
	CreatedAt time.Time
}

type RabbitMQProducer struct {
	ch *amqp.Channel
}

const AgentResponseReceivedEventTopic = "agent_response_received"

func NewRabbitMQProducer(ch *amqp.Channel) AgentMsgProducer {
	return &RabbitMQProducer{ch: ch}
}

func (s *RabbitMQProducer) AgentProduceResponseReceivedEvent(evt ResponseReceivedEvent) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = s.ch.PublishWithContext(ctx,
		"",                              // exchange
		AgentResponseReceivedEventTopic, // routing key
		false,                           // mandatory
		false,                           // immediate
		amqp.Publishing{
			ContentType: "json/application",
			Body:        []byte(body),
		})
	log.Printf(" [x] Sent %s\n", body)
	return err
}
