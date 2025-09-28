package repository

type InMemoryRepository struct {
    storage struct {
        i map[string]int64
        f map[string]float64
    }
}

func NewInMemoryRepo() *InMemoryRepository {
    return &InMemoryRepository{
        storage: struct {
            i map[string]int64
            f map[string]float64
        }{i: make(map[string]int64), f: make(map[string]float64)},
    }
}

func (r InMemoryRepository) GetAllFloat() (map[string]float64, error) {
    return r.storage.f, nil
}

func (r InMemoryRepository) GetAllInt() (map[string]int64, error) {
    return r.storage.i, nil
}

func (r InMemoryRepository) GetFloat(name string) (float64, error) {
    v, ok := r.storage.f[name]
    if !ok {
        return 0.0, ErrValueNotFound
    }
    return v, nil
}

func (r InMemoryRepository) GetInt(name string) (int64, error) {
    v, ok := r.storage.i[name]
    if !ok {
        return 0, ErrValueNotFound
    }
    return v, nil
}

func (r InMemoryRepository) SetFloat(name string, value float64) error {
    r.storage.f[name] = value
    return nil
}

func (r InMemoryRepository) SetInt(name string, value int64) error {
    r.storage.i[name] = value
    return nil
}
