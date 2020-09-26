// Package producer allows to submit new messages to a message broker.
package producer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

func New(address, topic string) *producer {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{address},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	return &producer{w}
}

type producer struct {
	w *kafka.Writer
}

func (p *producer) Produce(ctx context.Context, key, value []byte) error {
	if err := p.w.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	}); err != nil {
		return err
	}

	return nil
}

func (p *producer) Close() error {
	return p.w.Close()
}
