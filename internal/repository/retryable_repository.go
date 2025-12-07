package repository

import (
    "context"
    "fmt"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "time"
)

type retryableRepository struct {
    repo MetricRepository
}

func (r retryableRepository) GetAll(ctx context.Context) ([]models.Metrics, error) {
    var result []models.Metrics

    err := performOperationWithRetry(
        func() error {
            res, err := r.repo.GetAll(ctx)
            if err == nil {
                result = res
                return nil
            }

            return err
        },
        func(err error) bool {
            return IsRetriable(err)
        },
    )

    if err != nil {
        return nil, err
    }

    return result, nil
}

func (r retryableRepository) Get(ctx context.Context, mID string, mType string) (*models.Metrics, error) {
    var result *models.Metrics

    err := performOperationWithRetry(
        func() error {
            res, err := r.repo.Get(ctx, mID, mType)
            if err == nil {
                result = res
                return nil
            }

            return err
        },
        func(err error) bool {
            return IsRetriable(err)
        },
    )

    return result, err
}

func (r retryableRepository) Set(ctx context.Context, metric models.Metrics) error {
    err := performOperationWithRetry(
        func() error {
            return r.repo.Set(ctx, metric)
        },
        func(err error) bool {
            return IsRetriable(err)
        },
    )

    if err != nil {
        return fmt.Errorf("failed to perform set: %w", err)
    }

    return nil
}

func (r retryableRepository) SetBatch(ctx context.Context, metrics []models.Metrics) error {
    err := performOperationWithRetry(
        func() error {
            return r.repo.SetBatch(ctx, metrics)
        },
        func(err error) bool {
            return IsRetriable(err)
        },
    )

    if err != nil {
        return fmt.Errorf("failed to perform set batch: %w", err)
    }

    return nil
}

func performOperationWithRetry(
    op func() error,
    shouldRetry func(err error) bool,
) error {
    err := op()
    if err == nil {
        return nil
    }

    if !shouldRetry(err) {
        return err
    }

    originalErr := err

    for a := 0; a <= 2; a++ {
        time.Sleep(time.Duration(1+2*a) * time.Second)
        err = op()
        if err == nil {
            return nil
        }

        if !shouldRetry(err) {
            return fmt.Errorf("failed to retry, original: %w", originalErr)
        }
    }

    return fmt.Errorf("failed to retry, tried 3 times. original err: %w", originalErr)
}
