package service

import (
    "github.com/stretchr/testify/require"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository/mocks"
    "log"
    "testing"
)

func Test_UpdateCounter_SaveNewValue(t *testing.T) {
    // Given
    expectedValue := int64(900)
    metricName := `test_metric`

    repo := mocks.NewMockRepository()
    service := makeService(repo)

    // When
    err := service.UpdateCounter(metricName, expectedValue)

    // Then
    require.NoError(t, err, `expected successful operation`)
    require.Equal(t, expectedValue, int64(repo.SetLastCallNameParam.Value))
}

func Test_UpdateCounter_UpdateExistingValue(t *testing.T) {
    // Given
    expectedValue := 1800
    storedValue := int64(900)
    metricName := `test_metric`

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetReturnValue = &repository.Metric{
        MetricName: metricName,
        TypeName:   CounterMetricTypeName,
        Value:      float64(storedValue),
    }

    // When
    err := service.UpdateCounter(metricName, storedValue)

    // Then
    require.NoError(t, err, `expected successful operation`)
    require.Equal(t, expectedValue, int(repo.SetLastCallNameParam.Value))
}

func Test_UpdateCounter_ExistingMetricWrongType_ThrowError(t *testing.T) {
    // Given
    storedValue := int64(900)
    metricName := `test_metric`

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetReturnValue = &repository.Metric{
        MetricName: metricName,
        TypeName:   GaugeMetricTypeName,
        Value:      float64(storedValue),
    }

    // When
    err := service.UpdateCounter(metricName, storedValue)

    // Then
    require.ErrorIs(t, err, ErrMetricTypeMismatch)
}

func Test_UpdateGauge_SaveNewValue(t *testing.T) {
    // Given
    expectedValue := 900.5
    metricName := `test_metric`

    repo := mocks.NewMockRepository()
    service := makeService(repo)

    // When
    err := service.UpdateGauge(metricName, expectedValue)

    // Then
    require.NoError(t, err, `expected successful operation`)
    require.Equal(t, expectedValue, repo.SetLastCallNameParam.Value)
}

func Test_UpdateGauge_UpdateExistingValue(t *testing.T) {
    // Given
    storedValue := 200.0
    expectedValue := 1800.0
    metricName := `test_metric`

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetReturnValue = &repository.Metric{
        MetricName: metricName,
        TypeName:   GaugeMetricTypeName,
        Value:      storedValue,
    }

    // When
    err := service.UpdateGauge(metricName, expectedValue)

    // Then
    require.NoError(t, err, `expected successful operation`)
    require.Equal(t, expectedValue, repo.SetLastCallNameParam.Value)
}

func Test_UpdateGauge_ExistingMetricWrongType_ThrowError(t *testing.T) {
    // Given
    storedValue := 200.0
    expectedValue := 1800.0
    metricName := `test_metric`

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetReturnValue = &repository.Metric{
        MetricName: metricName,
        TypeName:   CounterMetricTypeName,
        Value:      storedValue,
    }

    // When
    err := service.UpdateGauge(metricName, expectedValue)

    // Then
    require.ErrorIs(t, err, ErrMetricTypeMismatch)
    require.False(t, repo.IsSetCalled, `no write operations expected`)
}

func makeService(repo repository.IMetricRepository) *MetricService {
    return &MetricService{
        log:  log.Default(),
        repo: repo,
    }
}
