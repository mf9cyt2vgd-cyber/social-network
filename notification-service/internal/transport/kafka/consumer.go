package kafka

import (
	"context"
	"log/slog"
	"time"

	"notification-service/internal/domain"
	"notification-service/internal/mapper"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type KafkaConsumer struct {
	subscriber *kafka.Subscriber
	log        *slog.Logger
	topic      string
	cancel     context.CancelFunc
	done       chan struct{}
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	topic string,
	log *slog.Logger,
) (*KafkaConsumer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Return.Errors = true

	subscriberConfig := kafka.SubscriberConfig{
		OverwriteSaramaConfig: saramaCfg,
		Brokers:               brokers,
		ConsumerGroup:         groupID,
		Unmarshaler:           kafka.DefaultMarshaler{},
		Tracer:                kafka.NewOTELSaramaTracer(),
		ReconnectRetrySleep:   5 * time.Second,
	}

	subscriber, err := kafka.NewSubscriber(subscriberConfig, watermill.NewSlogLogger(log))
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		subscriber: subscriber,
		log:        log,
		topic:      topic,
	}, nil
}

func (k *KafkaConsumer) Consume(ctx context.Context) chan *domain.PostEvent {
	out := make(chan *domain.PostEvent)
	k.done = make(chan struct{})
	consumeCtx, cancel := context.WithCancel(ctx)
	k.cancel = cancel

	go func() {
		defer close(out)
		defer close(k.done)

		messages, err := k.subscriber.Subscribe(ctx, k.topic)
		if err != nil {
			k.log.Error("failed to subscribe to topic", "error", err)
			return
		}
		for msg := range messages {
			func(m *message.Message) {
				propagator := otel.GetTextMapPropagator()
				parentCtx := propagator.Extract(m.Context(), propagation.MapCarrier(m.Metadata))

				tr := otel.Tracer("notification-service")
				childCtx, span := tr.Start(parentCtx, "notification-process")
				defer span.End()
				post, err := mapper.ConvertKafkaMessageIntoPost(m.Payload)
				if err != nil {
					k.log.Error("failed to convert Kafka message", err)
					return
				}
				select {
				case out <- &domain.PostEvent{
					Post:     &post,
					TraceCtx: childCtx,
				}:
					msg.Ack()
				case <-consumeCtx.Done():
					span.RecordError(consumeCtx.Err())
					span.End()
					msg.Nack()
					return
				}
			}(msg)
		}
	}()

	return out
}

func (k *KafkaConsumer) Close() error {
	if k.cancel != nil {
		k.cancel()
	}
	return k.subscriber.Close()
}
