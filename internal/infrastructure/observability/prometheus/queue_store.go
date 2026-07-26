package prometheus

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/postgresql"
	"github.com/AlekseyPorandaykin/go-kit/pkg/metrics"
)

var _ postgresql.QueueStore = (*QueueStore)(nil)

type QueueStore struct {
	inner postgresql.QueueStore
	db    string
}

func NewQueueStore(inner postgresql.QueueStore, db string) *QueueStore {
	return &QueueStore{inner: inner, db: db}
}

func (s *QueueStore) Publish(ctx context.Context, messages ...postgresql.QueueMessage) error {
	defer metrics.DBQueryHelper(s.db, "publish_queue")()
	return s.inner.Publish(ctx, messages...)
}

func (s *QueueStore) Receive(ctx context.Context, name string, limit int) ([]postgresql.QueueMessage, error) {
	defer metrics.DBQueryHelper(s.db, "receive_queue")()
	return s.inner.Receive(ctx, name, limit)
}

func (s *QueueStore) Delete(ctx context.Context, ids []string) error {
	defer metrics.DBQueryHelper(s.db, "delete_queue")()
	return s.inner.Delete(ctx, ids)
}

func (s *QueueStore) DeleteExpired(ctx context.Context, name string, before time.Time) error {
	return s.inner.DeleteExpired(ctx, name, before)
}
