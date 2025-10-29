package mocks

import (
    "context"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
)

type MetricServiceMock struct {
    IsGetAllCalled    bool
    GetAllReturnValue []models.Metrics
    GetAllReturnError error

    IsGetCalled             bool
    GetLastCallParam        models.Metrics
    GetReturnValue          bool
    GetReturnError          error
    GetLastCallParamMutator func(metrics *models.Metrics)

    IsUpdateCounterCalled           bool
    UpdateCounterLastCallNameParam  string
    UpdateCounterLastCallValueParam int64
    UpdateCounterReturnError        error

    IsUpdateGaugeCalled           bool
    UpdateGaugeLastCallNameParam  string
    UpdateGaugeLastCallValueParam float64
    UpdateGaugeReturnError        error

    IsSetBatchCalled      bool
    SetBatchLastCallParam []models.Metrics
    SetBatchReturnError   error

    IsPingCalled    bool
    PingReturnError error
}

func NewServiceMock() *MetricServiceMock {
    return &MetricServiceMock{}
}

func (m *MetricServiceMock) GetAll(ctx context.Context) ([]models.Metrics, error) {
    m.IsGetAllCalled = true
    return m.GetAllReturnValue, m.GetAllReturnError
}

func (m *MetricServiceMock) Get(ctx context.Context, metric *models.Metrics) (bool, error) {
    m.IsGetCalled = true
    m.GetLastCallParam = *metric

    if m.GetLastCallParamMutator != nil {
        m.GetLastCallParamMutator(metric)
    }

    return m.GetReturnValue, m.GetReturnError
}

func (m *MetricServiceMock) UpdateCounter(ctx context.Context, mID string, value int64) error {
    m.IsUpdateCounterCalled = true
    m.UpdateCounterLastCallNameParam = mID
    m.UpdateCounterLastCallValueParam = value
    return m.UpdateCounterReturnError
}

func (m *MetricServiceMock) UpdateGauge(ctx context.Context, mID string, value float64) error {
    m.IsUpdateGaugeCalled = true
    m.UpdateGaugeLastCallNameParam = mID
    m.UpdateGaugeLastCallValueParam = value
    return m.UpdateGaugeReturnError
}

func (m *MetricServiceMock) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
    m.IsSetBatchCalled = true
    m.SetBatchLastCallParam = metrics
    return m.SetBatchReturnError
}

func (m *MetricServiceMock) Ping(ctx context.Context) error {
    m.IsPingCalled = true
    return m.PingReturnError
}
