package forwarder

import (
	"context"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	wmsql "github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	KafkaBrokers []string
	Logger       *slog.Logger
}

type Forwarder struct {
	fwd    *forwarder.Forwarder
	cancel context.CancelFunc
	done   chan struct{}
}

func StartForwarder(ctx context.Context, pool *pgxpool.Pool, cfg Config) (*Forwarder, error) {
	logger := watermill.NewSlogLogger(cfg.Logger)
	db := wmsql.BeginnerFromPgx(pool)
	fwdCtx, cancel := context.WithCancel(ctx)
	schemaAdapter := wmsql.DefaultPostgreSQLSchema{GenerateMessagesTableName: func(topic string) string {
		return "posts_outbox"
	}}

	sqlSub, err := wmsql.NewSubscriber(
		db,
		wmsql.SubscriberConfig{
			OffsetsAdapter:   wmsql.DefaultPostgreSQLOffsetsAdapter{},
			SchemaAdapter:    schemaAdapter,
			PollInterval:     300 * time.Millisecond,
			InitializeSchema: false,
			RetryInterval:    1 * time.Second},
		logger,
	)
	if err != nil {
		cancel()
		return nil, err
	}

	kafkaPub, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   cfg.KafkaBrokers,
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger)
	if err != nil {
		cancel()
		return nil, err
	}
	fwd, err := forwarder.NewForwarder(sqlSub, kafkaPub, logger, forwarder.Config{ForwarderTopic: "posts_outbox_forwarder"})
	if err != nil {
		cancel()
		return nil, err
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		defer cancel()

		if err := fwd.Run(fwdCtx); err != nil && err != context.Canceled {
			logger.Error("forwarder stopped with error", err, nil)
		}

	}()

	return &Forwarder{
		fwd:    fwd,
		cancel: cancel,
		done:   done,
	}, nil
}
func (f *Forwarder) Stop() error {
	f.cancel()
	select {
	case <-f.done:
	case <-time.After(10 * time.Second):
	}

	if err := f.fwd.Close(); err != nil {
		return err
	}
	return nil
}
