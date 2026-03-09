package postgres

import (
	"context"
	"encoding/json"
	"log/slog"
	"post-service/internal/domain"

	"github.com/ThreeDotsLabs/watermill"
	wmsql "github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPostRepository struct {
	pool      *pgxpool.Pool
	publisher message.Publisher
	logger    *slog.Logger
}
type txKey struct {
}

func NewPostgresPostRepository(pool *pgxpool.Pool, log *slog.Logger) (*PostgresPostRepository, error) {
	db := wmsql.BeginnerFromPgx(pool)

	outboxPublisher, err := wmsql.NewPublisher(
		db,
		wmsql.PublisherConfig{
			SchemaAdapter: wmsql.DefaultPostgreSQLSchema{GenerateMessagesTableName: func(topic string) string {
				return "posts_outbox"
			}},
		},
		watermill.NewSlogLogger(log))
	if err != nil {
		return nil, err
	}
	publisher := forwarder.NewPublisher(outboxPublisher, forwarder.PublisherConfig{
		ForwarderTopic: "posts_outbox_forwarder",
	})
	return &PostgresPostRepository{pool: pool, publisher: publisher, logger: log}, nil
}

func (r *PostgresPostRepository) Save(ctx context.Context, post *domain.Post) error {
	tx, err := r.pool.Begin(ctx)

	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO posts (id, title, author, content, tags, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		post.ID, post.Title, post.Author, post.Content, post.Tags, post.CreatedAt)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]any{
		"id":         post.ID,
		"title":      post.Title,
		"author":     post.Author,
		"content":    post.Content,
		"tags":       post.Tags,
		"created_at": post.CreatedAt,
	})
	if err != nil {
		return err
	}
	var txContextKey txKey
	txCtx := context.WithValue(ctx, txContextKey, tx)
	msg := message.NewMessage(
		watermill.NewUUID(),
		payload)

	msg.SetContext(txCtx)
	err = r.publisher.Publish("posts.created", msg)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresPostRepository) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, title, author, content, tags, created_at FROM posts WHERE id=$1`, id)

	var p domain.Post
	err := row.Scan(&p.ID, &p.Title, &p.Author, &p.Content, &p.Tags, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresPostRepository) List(ctx context.Context) ([]*domain.Post, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, author, content, tags, created_at FROM posts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Author, &p.Content, &p.Tags, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, &p)
	}
	return posts, nil
}
func (r *PostgresPostRepository) Close() error {
	if r.publisher != nil {
		return r.publisher.Close()
	}
	return nil
}
