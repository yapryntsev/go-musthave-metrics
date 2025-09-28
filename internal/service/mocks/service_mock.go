package mocks

type MetricServiceMock struct {
    IsGetAllCalled    bool
    GetAllReturnValue map[string]string
    GetAllReturnError error

    IsGetCalled                bool
    GetLastCallMetricTypeParam string
    GetLastCallNameParam       string
    GetReturnValue             string
    GetReturnError             error

    IsUpdateCounterCalled           bool
    UpdateCounterLastCallNameParam  string
    UpdateCounterLastCallValueParam int64
    UpdateCounterReturnError        error

    IsUpdateGaugeCalled           bool
    UpdateGaugeLastCallNameParam  string
    UpdateGaugeLastCallValueParam float64
    UpdateGaugeReturnError        error
}

func NewServiceMock() *MetricServiceMock {
    return &MetricServiceMock{}
}

func (m *MetricServiceMock) GetAll() (map[string]string, error) {
    m.IsGetAllCalled = true
    return m.GetAllReturnValue, m.GetAllReturnError
}

func (m *MetricServiceMock) Get(metricType string, name string) (string, error) {
    m.IsGetCalled = true
    m.GetLastCallNameParam = name
    m.GetLastCallMetricTypeParam = metricType
    return m.GetReturnValue, m.GetReturnError
}

func (m *MetricServiceMock) UpdateCounter(name string, value int64) error {
    m.IsUpdateCounterCalled = true
    m.UpdateCounterLastCallNameParam = name
    m.UpdateCounterLastCallValueParam = value
    return m.UpdateCounterReturnError
}

func (m *MetricServiceMock) UpdateGauge(name string, value float64) error {
    m.IsUpdateGaugeCalled = true
    m.UpdateGaugeLastCallNameParam = name
    m.UpdateGaugeLastCallValueParam = value
    return m.UpdateGaugeReturnError
}
