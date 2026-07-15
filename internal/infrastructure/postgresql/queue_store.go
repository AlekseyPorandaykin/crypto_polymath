package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type QueueMessage struct {
	ID        string    `db:"id"`
	KeyEvent  string    `db:"key_event"`
	Name      string    `db:"name"`
	Body      []byte    `db:"body"`
	CreatedAt time.Time `db:"created_at"`
}

type QueueStore interface {
	Publish(ctx context.Context, messages ...QueueMessage) error
	Receive(ctx context.Context, name string, limit int) ([]QueueMessage, error)
	Delete(ctx context.Context, ids []string) error
	DeleteExpired(ctx context.Context, name string, before time.Time) error
}

type PGQueueStore struct {
	db *sqlx.DB
}

func NewQueueStore(db *sqlx.DB) *PGQueueStore {
	return &PGQueueStore{db: db}
}

func (s *PGQueueStore) Publish(ctx context.Context, messages ...QueueMessage) error {
	query := `
INSERT INTO crypto_polymath.queues(id, key_event, name, body, created_at)
VALUES 
`
	valuesQuery := `(:id, :key_event, :name, :body, :created_at)`
	valuesQueries := make([]string, 0, len(messages))
	values := make([]any, 0, len(messages))
	for i := range messages {
		preparedQueryV, vals, err := s.db.BindNamed(valuesQuery, messages[i])
		if err != nil {
			return err
		}
		valuesQueries = append(valuesQueries, preparedQueryV)
		values = append(values, vals...)
	}
	preparedQuery := fmt.Sprintf("%s %s", query, strings.Join(valuesQueries, ","))
	_, err := s.db.ExecContext(ctx, preparedQuery, values...)
	if err != nil {
		return fmt.Errorf("failed to publish messages: %w", err)
	}
	return nil
}

func (s *PGQueueStore) Receive(ctx context.Context, name string, limit int) ([]QueueMessage, error) {
	query := `
SELECT id, key_event, name, body, created_at
FROM crypto_polymath.queues
WHERE name = $1
ORDER BY created_at ASC
LIMIT $2
`
	messages := make([]QueueMessage, 0, limit)
	if err := s.db.SelectContext(ctx, &messages, query, name, limit); err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}
	return messages, nil
}

func (s *PGQueueStore) Delete(ctx context.Context, ids []string) error {
	query := `
DELETE FROM crypto_polymath.queues WHERE id IN (?)
`
	preparedQuery, values, err := sqlx.In(query, ids)
	if err != nil {
		return fmt.Errorf("failed to prepare delete query: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, s.db.Rebind(preparedQuery), values...); err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	return nil
}

func (s *PGQueueStore) DeleteExpired(ctx context.Context, name string, before time.Time) error {
	query := `
DELETE FROM crypto_polymath.queues WHERE name = $1 AND created_at < $2
`
	_, err := s.db.ExecContext(ctx, query, name, before)
	if err != nil {
		return err
	}
	return nil
}
