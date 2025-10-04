package repository

type MemoryMetricRepository struct {
    storage struct {
        i map[string]int64
        f map[string]float64
    }
}

func NewInMemoryRepo() *MemoryMetricRepository {
    return &MemoryMetricRepository{
        storage: struct {
            i map[string]int64
            f map[string]float64
        }{i: make(map[string]int64), f: make(map[string]float64)},
    }
}

func (r MemoryMetricRepository) GetAllFloat() (map[string]float64, error) {
    return r.storage.f, nil
}

func (r MemoryMetricRepository) GetAllInt() (map[string]int64, error) {
    return r.storage.i, nil
}

func (r MemoryMetricRepository) GetFloat(name string) (float64, error) {
    v, ok := r.storage.f[name]
    if !ok {
        return 0.0, ErrValueNotFound
    }
    return v, nil
}

func (r MemoryMetricRepository) GetInt(name string) (int64, error) {
    v, ok := r.storage.i[name]
    if !ok {
        return 0, ErrValueNotFound
    }
    return v, nil
}

func (r MemoryMetricRepository) SetFloat(name string, value float64) error {
    r.storage.f[name] = value
    return nil
}

func (r MemoryMetricRepository) SetInt(name string, value int64) error {
    r.storage.i[name] = value
    return nil
}
