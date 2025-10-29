package service

import (
    "errors"
    "github.com/stretchr/testify/assert"
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
    err := service.UpdateCounter(t.Context(), mID, expectedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, models.Metrics{ID: mID, MType: models.Counter, Delta: &expectedValue}, repo.SetLastCallParam)
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
    err := service.UpdateCounter(t.Context(), mID, storedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, models.Metrics{ID: mID, MType: models.Counter, Delta: &expectedValue}, repo.SetLastCallParam)
}

func Test_UpdateGauge_SaveNewValue(t *testing.T) {
    // Given
    expectedValue := 900.5
    mID := "test_metric"

    repo := mocks.NewMockRepository()
    service := makeService(repo)

    // When
    err := service.UpdateGauge(t.Context(), mID, expectedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, models.Metrics{ID: mID, MType: models.Gauge, Value: &expectedValue}, repo.SetLastCallParam)
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
    err := service.UpdateGauge(t.Context(), mID, expectedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, models.Metrics{ID: mID, MType: models.Gauge, Value: &expectedValue}, repo.SetLastCallParam)
}

func Test_Get_GotErrorFromRepo_Rethrow(t *testing.T) {
    // Given
    expectedError := errors.New("test error")
    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetReturnError = expectedError

    // When
    value, err := service.Get(
        t.Context(),
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
    ok, err := service.Get(t.Context(), metric)

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
    value, err := service.GetAll(t.Context())

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
    repo.GetAllReturnValue = []models.Metrics{expectedValue}

    // When
    value, err := service.GetAll(t.Context())

    // Then
    require.True(t, repo.IsGetAllCalled, "expected to call repo for value")
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, []models.Metrics{expectedValue}, value)
}

func Test_UpdateBatch_EmptyBatch_DoNothing(t *testing.T) {
    // Given
    var batch []models.Metrics

    repo := mocks.NewMockRepository()
    service := makeService(repo)

    // When
    err := service.UpdateBatch(t.Context(), batch)

    // Then
    assert.False(t, repo.IsSetBatchCalled, "unexpected set batch call")
    assert.NoError(t, err, "expected successful operation")
}

func Test_UpdateBatch_ValidBatch_PassToRepo(t *testing.T) {
    // Given
    expectedValue := 10.0
    expectedDelta := int64(20)
    expectedBatch := []models.Metrics{
        {ID: "id-1", MType: models.Gauge, Value: new(float64)},
        {ID: "id-2", MType: models.Counter, Delta: new(int64)},
    }

    *expectedBatch[0].Value = expectedValue
    *expectedBatch[1].Delta = expectedDelta

    repo := mocks.NewMockRepository()
    service := makeService(repo)

    // When
    err := service.UpdateBatch(t.Context(), expectedBatch)

    // Then
    assert.True(t, repo.IsSetBatchCalled, "expected to pass batch to repo")
    assert.NoError(t, err, "expected successful operation")
    assert.ElementsMatch(t, repo.SetBatchLastCallParam, expectedBatch, "repo got wrong batch")
}

func Test_UpdateBatch_BatchContainsDuplicateCounter_MergeBeforePassToRepo(t *testing.T) {
    // Given
    expectedDelta := int64(10)
    expectedResult := int64(20)
    expectedBatch := []models.Metrics{
        {ID: "id-1", MType: models.Counter, Delta: new(int64)},
        {ID: "id-1", MType: models.Counter, Delta: new(int64)},
    }

    *expectedBatch[0].Delta = expectedDelta
    *expectedBatch[1].Delta = expectedDelta

    repo := mocks.NewMockRepository()
    service := makeService(repo)

    // When
    err := service.UpdateBatch(t.Context(), expectedBatch)

    // Then
    assert.True(t, repo.IsSetBatchCalled, "expected to pass batch to repo")
    assert.NoError(t, err, "expected successful operation")
    assert.Len(t, repo.SetBatchLastCallParam, 1, "expected to get 1 merged metric")
    assert.Equal(t, *repo.SetBatchLastCallParam[0].Delta, expectedResult)
}

func makeService(repo repository.MetricRepository) *Service {
    return &Service{
        repo: repo,
    }
}
