package mocks

import (
    "context"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
)

type MockRepository struct {
    IsGetAllCalled    bool
    GetAllReturnValue []models.Metrics
    GetAllReturnError error

    IsGetCalled            bool
    GetLastCallFirstParam  string
    GetLastCallSecondParam string
    GetReturnValue         *models.Metrics
    GetReturnError         error

    IsSetCalled      bool
    SetLastCallParam models.Metrics
    SetReturnError   error

    IsSetBatchCalled      bool
    SetBatchLastCallParam []models.Metrics
    SetBatchReturnError   error
}

func NewMockRepository() *MockRepository {
    return &MockRepository{}
}

func (m *MockRepository) GetAll(ctx context.Context,) ([]models.Metrics, error) {
    m.IsGetAllCalled = true
    return m.GetAllReturnValue, m.GetAllReturnError
}

func (m *MockRepository) Get(ctx context.Context, mID string, mType string) (*models.Metrics, error) {
    m.IsGetCalled = true
    m.GetLastCallFirstParam = mID
    m.GetLastCallSecondParam = mType
    return m.GetReturnValue, m.GetReturnError
}

func (m *MockRepository) Set(ctx context.Context, metric models.Metrics) error {
    m.IsSetCalled = true
    m.SetLastCallParam = metric
    return m.SetReturnError
}

func (m *MockRepository) SetBatch(ctx context.Context, metrics []models.Metrics) error {
    m.IsSetBatchCalled = true
    m.SetBatchLastCallParam = metrics
    return m.SetBatchReturnError
}
