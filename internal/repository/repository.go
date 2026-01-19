package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	models "github.com/yapryntsev/go-musthave-metrics/internal/model"
	"go.uber.org/zap"
)

// MetricRepository provides methods to store, retrieve, and update metrics in persistent storage.
type MetricRepository interface {
	// GetAll returns all collected metrics.
	GetAll(ctx context.Context) ([]models.Metrics, error)
	// Get retrieves a metric specified by its ID and type.
	Get(ctx context.Context, mID string, mType string) (*models.Metrics, error)
	// Set overrides an existing metric with a new one using the same ID.
	Set(ctx context.Context, metric models.Metrics) error
	// SetBatch works like Set but ensures transactional integrity for all provided metrics.
	SetBatch(ctx context.Context, metrics []*models.Metrics) error
}

// New returns new MetricRepository instance.
func New(dbPool *pgxpool.Pool, storeInt time.Duration, filePath string, restore bool, l *zap.Logger) MetricRepository {
	if dbPool != nil {
		return retryableRepository{
			repo: newDatabaseRepository(dbPool, l),
		}
	}

	return newFileRepository(storeInt, filePath, restore, l)
}
