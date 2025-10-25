package repository

import "sync"

type MemoryMetricRepository struct {
    mu      sync.RWMutex
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

func (r *MemoryMetricRepository) GetAllFloat() (map[string]float64, error) {
    r.mu.RLock()
    v := r.storage.f
    r.mu.RUnlock()

    return v, nil
}

func (r *MemoryMetricRepository) GetAllInt() (map[string]int64, error) {
    r.mu.RLock()
    v := r.storage.i
    r.mu.RUnlock()

    return v, nil
}

func (r *MemoryMetricRepository) GetFloat(name string) (float64, error) {
    r.mu.RLock()
    v, ok := r.storage.f[name]
    r.mu.RUnlock()

    if !ok {
        return 0.0, ErrValueNotFound
    }
    return v, nil
}

func (r *MemoryMetricRepository) GetInt(name string) (int64, error) {
    r.mu.RLock()
    v, ok := r.storage.i[name]
    r.mu.RUnlock()

    if !ok {
        return 0, ErrValueNotFound
    }
    return v, nil
}

func (r *MemoryMetricRepository) SetFloat(name string, value float64) error {
    r.mu.Lock()
    r.storage.f[name] = value
    r.mu.Unlock()

    return nil
}

func (r *MemoryMetricRepository) SetInt(name string, value int64) error {
    r.mu.Lock()
    r.storage.i[name] = value
    r.mu.Unlock()

    return nil
}
