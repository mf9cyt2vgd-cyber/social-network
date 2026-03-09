package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"notification-service/internal/domain"
	"notification-service/internal/mapper"
	"strings"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaConsumer struct {
	*kafka.Consumer
	log    *slog.Logger
	cancel context.CancelFunc
	done   chan struct{}
}

func NewKafkaConsumer(brokers []string, groupId string, topic string, log *slog.Logger) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka: %w", err)
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
func (k *KafkaConsumer) Consume(ctx context.Context) chan *domain.Post {
	consumeCtx, cancel := context.WithCancel(ctx)
	k.cancel = cancel
	k.done = make(chan struct{})
	out := make(chan *domain.Post)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(k.done)
		for {
			select {
			case <-consumeCtx.Done():
				k.log.Info("Got ctx signal", "signal", ctx.Err())
				return
			default:
				msg, err := k.ReadMessage(time.Second)
				if err != nil {
					if kErr, ok := err.(kafka.Error); ok && !kErr.IsTimeout() {
						k.log.Error("Kafka read failed", "error", err)
					}
					continue
				}

				received, err := mapper.ConvertKafkaMessageIntoPost(msg.Value)
				if err != nil {
					k.log.Error("failed to convert Kafka message", "error", err)
					continue
				}

				out <- &received
			}
		}
	}()
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
func (k *KafkaConsumer) Close() error {
	if k.cancel == nil {
		return fmt.Errorf("should start consumer before calling close func")
	}
	k.log.Info("got stop signal")
	k.cancel()
	err := k.Consumer.Close()
	if err != nil {
		return err
	}
	select {
	case <-k.done:
	case <-time.After(10 * time.Second):
	}
	k.log.Info("consumer stopped gracefully")
	return nil
}
