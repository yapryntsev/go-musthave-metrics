package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	models "github.com/yapryntsev/go-musthave-metrics/internal/model"
	"go.uber.org/zap"
)

type MetricRepository interface {
	GetAll(ctx context.Context) ([]models.Metrics, error)
	Get(ctx context.Context, mID string, mType string) (*models.Metrics, error)
	Set(ctx context.Context, metric models.Metrics) error
	SetBatch(ctx context.Context, metrics []*models.Metrics) error
}

func New(dbPool *pgxpool.Pool, storeInt time.Duration, filePath string, restore bool, l *zap.Logger) MetricRepository {
	if dbPool != nil {
		return retryableRepository{
			repo: newDatabaseRepository(dbPool, l),
		}
	}

	return newFileRepository(storeInt, filePath, restore, l)
}
