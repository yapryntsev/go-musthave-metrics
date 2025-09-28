package mocks

type MockRepository struct {
    IsGetAllFloatCalled    bool
    GetAllFloatReturnValue map[string]float64
    GetAllFloatReturnError error

    IsGetAllIntCalled    bool
    GetAllIntReturnValue map[string]int64
    GetAllIntReturnError error

    IsGetFloatCalled          bool
    GetFloatLastCallNameParam string
    GetFloatReturnValue       float64
    GetFloatReturnError       error

    IsGetIntCalled          bool
    GetIntLastCallNameParam string
    GetIntReturnValue       int64
    GetIntReturnError       error

    IsSetFloatCalled           bool
    SetFloatLastCallNameParam  string
    SetFloatLastCallValueParam float64
    SetFloatReturnError        error

    IsSetIntCalled           bool
    SetIntLastCallNameParam  string
    SetIntLastCallValueParam int64
    SetIntReturnError        error
}

func NewMockRepository() *MockRepository {
    return &MockRepository{}
}

func (m *MockRepository) GetAllFloat() (map[string]float64, error) {
    m.IsGetAllFloatCalled = true
    return m.GetAllFloatReturnValue, m.GetAllFloatReturnError
}

func (m *MockRepository) GetAllInt() (map[string]int64, error) {
    m.IsGetAllIntCalled = true
    return m.GetAllIntReturnValue, m.GetAllIntReturnError
}

func (m *MockRepository) GetFloat(name string) (float64, error) {
    m.IsGetFloatCalled = true
    m.GetFloatLastCallNameParam = name
    return m.GetFloatReturnValue, m.GetFloatReturnError
}

func (m *MockRepository) GetInt(name string) (int64, error) {
    m.IsGetIntCalled = true
    m.GetIntLastCallNameParam = name
    return m.GetIntReturnValue, m.GetIntReturnError
}

func (m *MockRepository) SetFloat(name string, value float64) error {
    m.IsSetFloatCalled = true
    m.SetFloatLastCallNameParam = name
    m.SetFloatLastCallValueParam = value
    return m.SetFloatReturnError
}

func (m *MockRepository) SetInt(name string, value int64) error {
    m.IsSetIntCalled = true
    m.SetIntLastCallNameParam = name
    m.SetIntLastCallValueParam = value
    return m.SetIntReturnError
}
