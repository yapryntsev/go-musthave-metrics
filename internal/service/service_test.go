package service

import (
    "errors"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository/mocks"
    "go.uber.org/mock/gomock"
    "testing"
)

func Test_UpdateCounter_SaveNewValue(t *testing.T) {
    // Given
    ctx := t.Context()

    expectedValue := int64(900)
    mID := "test_metric"

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        Set(ctx, models.Metrics{ID: mID, MType: models.Counter, Delta: &expectedValue}).
        Times(1)
    repo.EXPECT().
        Get(ctx, mID, models.Counter).
        Return(nil, nil)

    // When
    err := service.UpdateCounter(ctx, mID, expectedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
}

func Test_UpdateCounter_UpdateExistingValue(t *testing.T) {
    // Given
    ctx := t.Context()

    expectedValue := int64(1800)
    storedValue := int64(900)
    mID := "test_metric"
    expectedMetric := &models.Metrics{
        ID:    mID,
        MType: models.Counter,
        Delta: &storedValue,
    }

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        Set(ctx, models.Metrics{ID: mID, MType: models.Counter, Delta: &expectedValue}).
        Times(1)
    repo.EXPECT().
        Get(ctx, mID, models.Counter).
        Return(expectedMetric, nil)

    // When
    err := service.UpdateCounter(ctx, mID, storedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
}

func Test_UpdateGauge_SaveNewValue(t *testing.T) {
    // Given
    ctx := t.Context()

    expectedValue := 900.5
    mID := "test_metric"

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        Set(ctx, models.Metrics{ID: mID, MType: models.Gauge, Value: &expectedValue}).
        Times(1)

    // When
    err := service.UpdateGauge(ctx, mID, expectedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
}

func Test_UpdateGauge_OverrideExistingValue(t *testing.T) {
    // Given
    ctx := t.Context()

    expectedValue := 1800.0
    mID := "test_metric"
    expectedMetric := models.Metrics{
        ID:    mID,
        MType: models.Gauge,
        Value: &expectedValue,
    }

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        Set(ctx, expectedMetric).
        Times(1)

    repo.EXPECT().
        Get(ctx, gomock.Any(), gomock.Any()).
        MaxTimes(0)

    // When
    err := service.UpdateGauge(ctx, mID, expectedValue)

    // Then
    require.NoError(t, err, "expected successful operation")
}

func Test_Get_GotErrorFromRepo_Rethrow(t *testing.T) {
    // Given
    ctx := t.Context()

    expectedError := errors.New("test error")
    metric := &models.Metrics{
        ID:    "test",
        MType: models.Gauge,
    }

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        Get(ctx, metric.ID, metric.MType).
        Return(nil, expectedError)

    // When
    value, err := service.Get(ctx, metric)

    // Then
    require.Empty(t, value, "expected to throw error, return value instead")
    require.ErrorIs(t, err, expectedError, "return error type mismatch")
}

func Test_Get_HasStoredValue_Return(t *testing.T) {
    // Given
    ctx := t.Context()

    expectedValue := 200.0
    expectedMetric := models.Metrics{
        ID:    "test",
        MType: models.Gauge,
        Value: &expectedValue,
    }

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        Get(ctx, expectedMetric.ID, expectedMetric.MType).
        Return(&expectedMetric, nil)

    // When
    metric := &models.Metrics{
        ID:    expectedMetric.ID,
        MType: expectedMetric.MType,
    }
    ok, err := service.Get(ctx, metric)

    // Then
    require.True(t, ok, "result expected to be true")
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, expectedMetric.ID, metric.ID)
    require.Equal(t, expectedMetric.MType, metric.MType)
    require.Equal(t, *expectedMetric.Value, *metric.Value)
}

func Test_GetAll_GotErrorFromRepo_Rethrow(t *testing.T) {
    // Given
    ctx := t.Context()
    expectedError := errors.New("test err")

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        GetAll(ctx).
        Return([]models.Metrics{}, expectedError)

    // When
    value, err := service.GetAll(ctx)

    // Then
    require.Empty(t, value, "expected to throw error, return value instead")
    require.ErrorIs(t, err, expectedError, "return error type mismatch")

}

func Test_GetAll_HasStoredValue_Return(t *testing.T) {
    // Given
    ctx := t.Context()
    expectedValue := models.Metrics{
        ID:    "test",
        MType: models.Counter,
        Delta: new(int64),
    }

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        GetAll(ctx).
        Return([]models.Metrics{expectedValue}, nil)

    // When
    value, err := service.GetAll(ctx)

    // Then
    require.NoError(t, err, "expected successful operation")
    require.Equal(t, []models.Metrics{expectedValue}, value)
}

func Test_UpdateBatch_EmptyBatch_DoNothing(t *testing.T) {
    // Given
    ctx := t.Context()
    var batch []models.Metrics

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        SetBatch(ctx, gomock.Any()).
        MaxTimes(0)

    // When
    err := service.UpdateBatch(ctx, batch)

    // Then
    assert.NoError(t, err, "expected successful operation")
}

func Test_UpdateBatch_ValidBatch_PassToRepo(t *testing.T) {
    // Given
    ctx := t.Context()

    expectedValue := 10.0
    expectedDelta := int64(20)
    expectedBatch := []models.Metrics{
        {ID: "id-1", MType: models.Gauge, Value: new(float64)},
        {ID: "id-2", MType: models.Counter, Delta: new(int64)},
    }

    *expectedBatch[0].Value = expectedValue
    *expectedBatch[1].Delta = expectedDelta

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        Get(ctx, gomock.Any(), gomock.Any()).
        Return(nil, nil)

    repo.EXPECT().
        SetBatch(ctx, expectedBatch)

    // When
    err := service.UpdateBatch(ctx, expectedBatch)

    // Then
    assert.NoError(t, err, "expected successful operation")
}

func Test_UpdateBatch_BatchContainsDuplicateCounter_MergeBeforePassToRepo(t *testing.T) {
    // Given
    ctx := t.Context()

    expectedDelta := int64(10)
    expectedResult := int64(20)
    expectedBatch := []models.Metrics{
        {ID: "id-1", MType: models.Counter, Delta: new(int64)},
        {ID: "id-1", MType: models.Counter, Delta: new(int64)},
    }

    *expectedBatch[0].Delta = expectedDelta
    *expectedBatch[1].Delta = expectedDelta

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    repo := mocks.NewMockMetricRepository(ctrl)
    service := makeService(repo)

    repo.EXPECT().
        Get(ctx, gomock.Any(), gomock.Any()).
        Return(nil, nil)

    repo.EXPECT().
        SetBatch(ctx, []models.Metrics{{ID: "id-1", MType: models.Counter, Delta: &expectedResult}})

    // When
    err := service.UpdateBatch(ctx, expectedBatch)

    // Then
    assert.NoError(t, err, "expected successful operation")
}

func makeService(repo repository.MetricRepository) *Service {
    return &Service{
        repo: repo,
    }
}
