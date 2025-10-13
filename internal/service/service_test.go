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
    metricName := `test_metric`

    repo := mocks.NewMockRepository()
    service := makeService(repo)

    // When
    err := service.UpdateCounter(metricName, expectedValue)

    // Then
    require.NoError(t, err, `expected successful operation`)
    require.Equal(t, expectedValue, repo.SetIntLastCallValueParam)
}

func Test_UpdateCounter_UpdateExistingValue(t *testing.T) {
    // Given
    expectedValue := int64(1800)
    storedValue := int64(900)
    metricName := `test_metric`

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetIntReturnValue = storedValue

    // When
    err := service.UpdateCounter(metricName, storedValue)

    // Then
    require.NoError(t, err, `expected successful operation`)
    require.Equal(t, expectedValue, repo.SetIntLastCallValueParam)
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
    require.Equal(t, expectedValue, repo.SetFloatLastCallValueParam)
}

func Test_UpdateGauge_UpdateExistingValue(t *testing.T) {
    // Given
    storedValue := 200.0
    expectedValue := 1800.0
    metricName := `test_metric`

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetFloatReturnValue = storedValue

    // When
    err := service.UpdateGauge(metricName, expectedValue)

    // Then
    require.NoError(t, err, `expected successful operation`)
    require.Equal(t, expectedValue, repo.SetFloatLastCallValueParam)
}

func Test_Get_GotErrorFromRepo_Rethrow(t *testing.T) {
    // Given
    expectedError := errors.New(`test error`)
    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetFloatReturnError = expectedError

    // When
    value, err := service.Get(
        &models.Metrics{
            ID:    "test",
            MType: models.Gauge,
        },
    )

    // Then
    require.True(t, repo.IsGetFloatCalled, `expected to call repo for value`)
    require.Empty(t, value, `expected to throw error, return value instead`)
    require.ErrorIs(t, err, expectedError, `return error type mismatch`)
    require.Equal(t, repo.GetFloatLastCallNameParam, `test`, `metric name mismatch`)
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
    repo.GetFloatReturnValue = expectedValue

    // When
    metric := &models.Metrics{
        ID:    expectedMetric.ID,
        MType: models.Gauge,
    }
    ok, err := service.Get(metric)

    // Then
    require.True(t, ok, "result expected to be true")
    require.True(t, repo.IsGetFloatCalled, `expected to call repo for value`)
    require.NoError(t, err, `expected successful operation`)
    require.Equal(t, expectedMetric.ID, metric.ID)
    require.Equal(t, expectedMetric.MType, metric.MType)
    require.Equal(t, *expectedMetric.Value, *metric.Value)
    require.Nil(t, expectedMetric.Delta, "delta property expected to be nil")
    require.Equal(t, repo.GetFloatLastCallNameParam, expectedMetric.ID, `metric name mismatch`)
}

func Test_GetAll_GotErrorFromRepo_Rethrow(t *testing.T) {
    fErr := errors.New(`float err`)
    iErr := errors.New(`int err`)

    tests := [...]struct {
        name     string
        floatErr error
        intErr   error
        want     error
    }{
        {`get float error`, fErr, nil, fErr},
        {`get int error`, nil, iErr, iErr},
    }

    for _, tc := range tests {
        t.Run(
            tc.name, func(t *testing.T) {
                // Given
                repo := mocks.NewMockRepository()
                service := makeService(repo)
                repo.GetAllFloatReturnError = tc.floatErr
                repo.GetAllIntReturnError = tc.intErr

                // When
                value, err := service.GetAll()

                // Then
                require.Empty(t, value, `expected to throw error, return value instead`)
                require.ErrorIs(t, err, tc.want, `return error type mismatch`)
            },
        )
    }
}

func Test_GetAll_HasStoredValue_Return(t *testing.T) {
    // Given
    expectedValue := map[string]float64{
        `test`: 64.2,
    }
    expectedFormattedValue := map[string]string{
        `test`: `64.200`,
    }

    repo := mocks.NewMockRepository()
    service := makeService(repo)
    repo.GetAllFloatReturnValue = expectedValue

    // When
    value, err := service.GetAll()

    // Then
    require.True(t, repo.IsGetAllFloatCalled, `expected to call repo for value`)
    require.NoError(t, err, `expected successful operation`)
    require.Equal(t, expectedFormattedValue, value)
}

func makeService(repo repository.MetricRepository) *Service {
    return &Service{
        log:  log.NewEntry(log.New()),
        repo: repo,
    }
}
