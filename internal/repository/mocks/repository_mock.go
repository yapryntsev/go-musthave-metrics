package mocks

import (
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
)

type MockRepository struct {
    IsGetAllCalled    bool
    GetAllReturnValue []*models.Metrics
    GetAllReturnError error

    IsGetCalled            bool
    GetLastCallFirstParam  string
    GetLastCallSecondParam string
    GetReturnValue         *models.Metrics
    GetReturnError         error

    IsSetCalled      bool
    SetLastCallParam *models.Metrics
    SetReturnError   error
}

func NewMockRepository() *MockRepository {
    return &MockRepository{}
}

func (m *MockRepository) GetAll() ([]*models.Metrics, error) {
    m.IsGetAllCalled = true
    return m.GetAllReturnValue, m.GetAllReturnError
}

func (m *MockRepository) Get(mID string, mType string) (*models.Metrics, error) {
    m.IsGetCalled = true
    m.GetLastCallFirstParam = mID
    m.GetLastCallSecondParam = mType
    return m.GetReturnValue, m.GetReturnError
}

func (m *MockRepository) Set(metric *models.Metrics) error {
    m.IsSetCalled = true
    m.SetLastCallParam = metric
    return m.SetReturnError
}
