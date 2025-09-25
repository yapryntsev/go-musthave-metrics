package repository

type Metric struct {
    MetricName string
    TypeName   string
    Value      float64
}

type IMetricRepository interface {
    Get(name string) (*Metric, error)
    Set(metric *Metric) error
}
