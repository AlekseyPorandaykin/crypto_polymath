package logging

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/postgresql"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"go.uber.org/zap"
)

var _ postgresql.QueueStore = (*QueueStore)(nil)

type QueueStore struct {
	inner  postgresql.QueueStore
	logger *zap.Logger
	db     string
}

func NewQueueStore(inner postgresql.QueueStore, logger *zap.Logger, db string) *QueueStore {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &QueueStore{inner: inner, logger: logger, db: db}
}

func (s *QueueStore) Publish(ctx context.Context, messages ...postgresql.QueueMessage) error {
	defer s.log(ctx, "publish_queue")()
	return s.inner.Publish(ctx, messages...)
}

func (s *QueueStore) Receive(ctx context.Context, name string, limit int) ([]postgresql.QueueMessage, error) {
	defer s.log(ctx, "receive_queue")()
	return s.inner.Receive(ctx, name, limit)
}

func (s *QueueStore) Delete(ctx context.Context, ids []string) error {
	defer s.log(ctx, "delete_queue")()
	return s.inner.Delete(ctx, ids)
}

func (s *QueueStore) DeleteExpired(ctx context.Context, name string, before time.Time) error {
	return s.inner.DeleteExpired(ctx, name, before)
}

func (s *QueueStore) log(ctx context.Context, query string) func() {
	now := time.Now()
	return func() {
		s.logger.Debug("Execute query",
			zap.String("query", query),
			zap.String("db", s.db),
			zap.String("duration", time.Since(now).String()),
			zap.String("request_id", util.RequestIDFromContext(ctx)),
		)
	}
}
