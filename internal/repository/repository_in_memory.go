package repository

type InMemoryRepository struct {
    storage map[string]struct {
        typeName string
        value    float64
    }
}

func NewInMemoryRepo() *InMemoryRepository {
    return &InMemoryRepository{
        storage: make(
            map[string]struct {
                typeName string
                value    float64
            },
        ),
    }
}

func (r InMemoryRepository) Get(name string) (*Metric, error) {
    v, ok := r.storage[name]
    if !ok {
        return nil, nil
    }

    metric := &Metric{
        MetricName: name,
        TypeName:   v.typeName,
        Value:      v.value,
    }
    return metric, nil
}

func (r InMemoryRepository) Set(metric *Metric) error {
    r.storage[metric.MetricName] = struct {
        typeName string
        value    float64
    }{
        typeName: metric.TypeName,
        value:    metric.Value,
    }

    return nil
}
