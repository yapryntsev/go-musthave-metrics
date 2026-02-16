package pool

import (
	"fmt"
	"sync"
)

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	pool sync.Pool
}

func New[T Resettable](factory func() *T) Pool[T] {
	return Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

func (p *Pool[T]) Get() (T, error) {
	obj, ok := p.pool.Get().(T)
	if !ok {
		return *new(T), fmt.Errorf("failed to cast %T to %T", obj, *new(T))
	}

	return obj, nil
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
