package mocks

import "github.com/yapryntsev/go-musthave-metrics/internal/repository"

type MockRepository struct {
    IsGetCalled          bool
    GetLastCallNameParam string
    GetReturnValue       *repository.Metric
    GetReturnError       error

    IsSetCalled          bool
    SetLastCallNameParam *repository.Metric
    SetReturnError       error
}

func NewMockRepository() *MockRepository {
    return &MockRepository{}
}

func (m *MockRepository) Get(name string) (*repository.Metric, error) {
    m.IsGetCalled = true
    m.GetLastCallNameParam = name
    return m.GetReturnValue, m.GetReturnError
}

func (m *MockRepository) Set(metric *repository.Metric) error {
    m.IsSetCalled = true
    m.SetLastCallNameParam = metric
    return m.SetReturnError
}
