package service

import (
    "errors"
    log "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/require"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository/mocks"
    "testing"
)

func Test_UpdateCounter_SaveNewValue(t *testing.T) {
    // Given
    expectedValue := int64(900)
    mID := "test_metric"

    repo := mocks.NewMockRepository()
    service := makeService(repo)

    // When
    err := service.UpdateCounter(mID, expectedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, &models.Metrics{ID: mID, MType: models.Counter, Delta: &expectedValue}, repo.SetLastCallParam)
}

func Test_UpdateCounter_UpdateExistingValue(t *testing.T) {
    // Given
    expectedValue := int64(1800)
    storedValue := int64(900)
    mID := "test_metric"

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetReturnValue = &models.Metrics{
        ID:    mID,
        MType: models.Counter,
        Delta: &storedValue,
    }

    // When
    err := service.UpdateCounter(mID, storedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, &models.Metrics{ID: mID, MType: models.Counter, Delta: &expectedValue}, repo.SetLastCallParam)
}

func Test_UpdateGauge_SaveNewValue(t *testing.T) {
    // Given
    expectedValue := 900.5
    mID := "test_metric"

    repo := mocks.NewMockRepository()
    service := makeService(repo)

    // When
    err := service.UpdateGauge(mID, expectedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, &models.Metrics{ID: mID, MType: models.Gauge, Value: &expectedValue}, repo.SetLastCallParam)
}

func Test_UpdateGauge_OverrideExistingValue(t *testing.T) {
    // Given
    storedValue := 200.0
    expectedValue := 1800.0
    mID := "test_metric"

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetReturnValue = &models.Metrics{
        ID:    mID,
        MType: models.Gauge,
        Value: &storedValue,
    }

    // When
    err := service.UpdateGauge(mID, expectedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, &models.Metrics{ID: mID, MType: models.Gauge, Value: &expectedValue}, repo.SetLastCallParam)
}

func Test_Get_GotErrorFromRepo_Rethrow(t *testing.T) {
    // Given
    expectedError := errors.New("test error")
    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetReturnError = expectedError

    // When
    value, err := service.Get(
        &models.Metrics{
            ID:    "test",
            MType: models.Gauge,
        },
    )

    // Then
    require.True(t, repo.IsGetCalled, "expected to call repo for value")
    require.Empty(t, value, "expected to throw error, return value instead")
    require.ErrorIs(t, err, expectedError, "return error type mismatch")
    require.Equal(t, repo.GetLastCallFirstParam, "test", "metric name mismatch")
    require.Equal(t, repo.GetLastCallSecondParam, models.Gauge, "metric type mismatch")
}

func Test_Get_HasStoredValue_Return(t *testing.T) {
    // Given
    expectedValue := 200.0
    expectedMetric := models.Metrics{
        ID:    "test",
        MType: models.Gauge,
        Value: &expectedValue,
    }

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetReturnValue = &expectedMetric

    // When
    metric := &models.Metrics{
        ID:    expectedMetric.ID,
        MType: expectedMetric.MType,
    }
    ok, err := service.Get(metric)

    // Then
    require.True(t, ok, "result expected to be true")
    require.True(t, repo.IsGetCalled, "expected to call repo for value")
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, expectedMetric.ID, metric.ID)
    require.Equal(t, expectedMetric.MType, metric.MType)
    require.Equal(t, *expectedMetric.Value, *metric.Value)
    require.Nil(t, expectedMetric.Delta, "delta property expected to be nil")
    require.Equal(t, repo.GetLastCallFirstParam, expectedMetric.ID, "metric name mismatch")
    require.Equal(t, repo.GetLastCallSecondParam, models.Gauge, "metric type mismatch")
}

func Test_GetAll_GotErrorFromRepo_Rethrow(t *testing.T) {
    // Given
    expectedError := errors.New("test err")
    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetAllReturnError = expectedError

    // When
    value, err := service.GetAll()

    // Then
    require.Empty(t, value, "expected to throw error, return value instead")
    require.ErrorIs(t, err, expectedError, "return error type mismatch")

}

func Test_GetAll_HasStoredValue_Return(t *testing.T) {
    // Given
    expectedValue := models.Metrics{
        ID:    "test",
        MType: models.Counter,
        Delta: new(int64),
    }

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetAllReturnValue = []*models.Metrics{&expectedValue}

    // When
    value, err := service.GetAll()

    // Then
    require.True(t, repo.IsGetAllCalled, "expected to call repo for value")
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, []*models.Metrics{&expectedValue}, value)
}

func makeService(repo repository.MetricRepository) *Service {
    return &Service{
        log:  log.NewEntry(log.New()),
        repo: repo,
    }
}
