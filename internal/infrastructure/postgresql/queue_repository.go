package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
)

type QueueRepository[T any] struct {
	store     QueueStore
	name      string
	ttl       time.Duration
	batchSize int
}

func NewQueueRepository[T any](store QueueStore, name string, ttl time.Duration, batchSize int) *QueueRepository[T] {
	if batchSize < 1 {
		batchSize = 500
	}
	return &QueueRepository[T]{store: store, name: name, ttl: ttl, batchSize: batchSize}
}

func (repo *QueueRepository[T]) Publish(message ...T) error {
	data := make([]QueueMessage, 0, len(message))
	for _, msg := range message {
		body, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		keyEvent := ""
		if ke, ok := any(msg).(queue.KeyEvent); ok {
			keyEvent = ke.KeyEvent()
		}
		data = append(data, QueueMessage{
			ID:        uuid.NewString(),
			KeyEvent:  keyEvent,
			Name:      repo.name,
			Body:      body,
			CreatedAt: time.Now().In(time.UTC),
		})
	}

	return repo.store.Publish(context.Background(), data...)
}

func (repo *QueueRepository[T]) Consume(ctx context.Context, handler func(message T) error) error {
	data, err := repo.store.Receive(ctx, repo.name, repo.batchSize)
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
		if err := repo.store.Delete(ctx, ids); err != nil {
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

func (repo *QueueRepository[T]) Receive(parentCtx context.Context) ([]*T, error) {
	ctx, cancel := context.WithTimeout(parentCtx, time.Minute)
	defer cancel()
	if err := repo.store.DeleteExpired(ctx, repo.name, time.Now().In(time.UTC).Add(-repo.ttl)); err != nil {
		return nil, fmt.Errorf("failed to delete old messages: %w", err)
	}
	data := make([]QueueMessage, 0, repo.batchSize)
	err := backoff.Retry(func() error {
		var err error
		data, err = repo.store.Receive(ctx, repo.name, repo.batchSize)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(time.Minute)))
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	response := make([]*T, 0, len(data))
	ids := make([]string, 0, len(data))
	for i := range data {
		var msg T
		if errU := json.Unmarshal(data[i].Body, &msg); errU != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %w", errU)
		}
		ids = append(ids, data[i].ID)
		response = append(response, &msg)
	}
	if errD := repo.store.Delete(ctx, ids); errD != nil {
		return nil, fmt.Errorf("failed to delete message: %w", errD)
	}
	return response, nil
}
