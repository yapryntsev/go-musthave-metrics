package repository

type inMemoryRepository struct {
    storage map[string]struct {
        typeName string
        value    float64
    }
}

func (r inMemoryRepository) Get(name string) (*Metric, error) {
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

func (r inMemoryRepository) Set(metric *Metric) error {
    r.storage[metric.MetricName] = struct {
        typeName string
        value    float64
    }{
        typeName: metric.TypeName,
        value:    metric.Value,
    }

    return nil
}
