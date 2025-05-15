package service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type CacheAdapter interface {
	Name() string
	Rebuild(ctx context.Context) error
	TTL() time.Duration
}

type CacheManager struct {
	adapters    []CacheAdapter
	lastRebuild map[string]time.Time
}

func NewCacheManager(adapters ...CacheAdapter) *CacheManager {
	return &CacheManager{
		adapters:    adapters,
		lastRebuild: make(map[string]time.Time),
	}
}

func (m *CacheManager) Rebuild(ctx context.Context) error {
	for _, adapter := range m.adapters {
		lastRebuild, has := m.lastRebuild[adapter.Name()]
		if !has || time.Since(lastRebuild) > adapter.TTL() {
			start := time.Now()
			if err := adapter.Rebuild(ctx); err != nil {
				return errors.Wrap(err, fmt.Sprintf("rebuild cache=%s", adapter.Name()))
			}
			m.lastRebuild[adapter.Name()] = time.Now()
			zap.L().Debug(
				"Cache rebuilt",
				zap.String("cache", adapter.Name()),
				zap.String("duration", time.Since(start).String()),
			)
		}
	}
	return nil
}
