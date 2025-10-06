package mocks

type MetricServiceMock struct {
}

func (m *MetricServiceMock) UpdateCounter(name string, value int64) error {
    return nil
}

func (m *MetricServiceMock) UpdateGauge(name string, value float64) error {
    return nil
}
