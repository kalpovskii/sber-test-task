package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) SendEvent(action string) {
	msg := kafka.Message{
		Key:   []byte(time.Now().Format(time.RFC3339Nano)),
		Value: []byte(action),
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(context.Background(), msg); err != nil {
		log.Println("failed to write kafka message:", err)
	}
}