package repository

import "errors"

var ErrValueNotFound = errors.New(`metric value not found`)

type IMetricRepository interface {
    GetAllFloat() (map[string]float64, error)
    GetAllInt() (map[string]int64, error)

    GetFloat(name string) (float64, error)
    GetInt(name string) (int64, error)

    SetFloat(name string, value float64) error
    SetInt(name string, value int64) error
}
