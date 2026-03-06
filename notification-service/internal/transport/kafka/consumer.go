package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaConsumer struct {
	*kafka.Consumer
	log *slog.Logger
}

func NewKafkaConsumer(brokers []string, groupId string, topic string, log *slog.Logger) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}
	err = c.Subscribe(topic, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe topic: %w", err)
	}
	return &KafkaConsumer{
		Consumer: c,
		log:      log,
	}, nil
}
func (k *KafkaConsumer) Consume(ctx context.Context, c chan []byte, cErr chan error) {
	for {
		select {
		case <-ctx.Done():
			close(c)
			close(cErr)
			k.log.Info("Got ctx signal", "signal", ctx.Err())
			return
		default:
			msg, err := k.ReadMessage(time.Second)
			if err == nil {
				c <- msg.Value
			} else if !err.(kafka.Error).IsTimeout() {
				k.log.Error("Kafka read failed", "error", err)
				cErr <- err
			}
		}

	}
}
