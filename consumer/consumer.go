// Package consumer allows to consume messages from a message broker.
package consumer

import (
	"context"
	"encoding/json"

	"github.com/orlangure/testing-kafka-in-go/reporter"
	"github.com/segmentio/kafka-go"
)

func New(address, topic string) *consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{address},
		Topic:   topic,
	})
	return &consumer{r}
}

type consumer struct {
	r *kafka.Reader
}

func (c *consumer) Consume(ctx context.Context) <-chan reporter.Event {
	messages := make(chan reporter.Event)

	go func() {
		defer close(messages)

		for {
			m, err := c.r.ReadMessage(ctx)
			if err != nil {
				return
			}

			var e reporter.Event

			if err := json.Unmarshal(m.Value, &e); err != nil {
				return
			}

			e.Type = string(m.Key)

			messages <- e
		}
	}()

	return messages
}

func (c *consumer) Close() error {
	return c.r.Close()
}
