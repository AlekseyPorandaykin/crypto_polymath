package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

type QueueMessage struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Body      []byte    `db:"body"`
	CreatedAt time.Time `db:"created_at"`
}

type QueueRepository[T any] struct {
	db   *sqlx.DB
	name string
}

func NewQueueRepository[T any](db *sqlx.DB, name string) *QueueRepository[T] {
	return &QueueRepository[T]{db: db, name: name}
}

func (repo *QueueRepository[T]) Publish(message ...T) error {
	data := make([]QueueMessage, 0, len(message))
	for _, msg := range message {
		body, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		data = append(data, QueueMessage{
			ID:        uuid.NewString(),
			Name:      repo.name,
			Body:      body,
			CreatedAt: time.Now().In(time.UTC),
		})
	}

	return repo.publish(context.Background(), data...)
}

func (repo *QueueRepository[T]) Consume(ctx context.Context, handler func(message T) error) error {
	data, err := repo.receive(ctx, repo.name, 100)
	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}
	ids := make([]string, 0, len(data))
	for _, item := range data {
		var msg T
		if err := json.Unmarshal(item.Body, &msg); err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}
		if err := handler(msg); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}
		ids = append(ids, item.ID)
	}
	if len(ids) > 0 {
		if err := repo.delete(ctx, ids, len(ids)); err != nil {
			return fmt.Errorf("failed to delete messages: %w", err)
		}
	}
	return nil
}

func (repo *QueueRepository[T]) Listen() {
	return
}

func (repo *QueueRepository[T]) Close() {
	return
}

func (repo *QueueRepository[T]) Receive() (*T, error) {
	data, err := repo.receive(context.Background(), repo.name, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	var msg T
	if err := json.Unmarshal(data[0].Body, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	if err := repo.delete(context.Background(), []string{data[0].ID}, 1); err != nil {
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}
	return &msg, nil
}

func (repo *QueueRepository[T]) publish(ctx context.Context, messages ...QueueMessage) error {
	query := `
INSERT INTO crypto_polymath.queues(id, name, body, created_at)
VALUES 
`
	valuesQuery := `(:id, :name, :body, :created_at)`
	valuesQueries := make([]string, 0, len(messages))
	values := make([]any, 0, len(messages))
	for i := range messages {
		preparedQueryV, vals, err := repo.db.BindNamed(valuesQuery, messages[i])
		if err != nil {
			return err
		}
		valuesQueries = append(valuesQueries, preparedQueryV)
		values = append(values, vals...)
	}
	preparedQuery := fmt.Sprintf("%s %s", query, strings.Join(valuesQueries, ","))
	_, err := repo.db.ExecContext(ctx, preparedQuery, values...)
	if err != nil {
		return fmt.Errorf("failed to publish messages: %w", err)
	}
	return nil
}

func (repo *QueueRepository[T]) receive(ctx context.Context, name string, limit int) ([]QueueMessage, error) {
	query := `
SELECT id, name, body, created_at
FROM crypto_polymath.queues
WHERE name = $1
ORDER BY created_at DESC
LIMIT $2
`
	messages := make([]QueueMessage, 0, limit)
	if err := repo.db.SelectContext(ctx, &messages, query, name, limit); err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}
	return messages, nil
}

func (repo *QueueRepository[T]) delete(ctx context.Context, ids []string, limit int) error {
	query := `
DELETE FROM crypto_polymath.queues WHERE id IN (?)
`
	preparedQuery, values, err := sqlx.In(query, ids)
	if err != nil {
		return fmt.Errorf("failed to prepare delete query: %w", err)
	}
	if _, err := repo.db.ExecContext(ctx, repo.db.Rebind(preparedQuery), values...); err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	return nil
}
