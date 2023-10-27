package pool

import (
	"errors"
	"sync"
)

type Resourcer[T any] interface {
	Create() (*T, error)
	Reset() error
}

type Pool[T any] struct {
	idle      []*T
	occupied  []*T
	capacity  uint8
	resourcer Resourcer[T]
	mu        sync.Mutex
}

func NewPool[T any](capacity uint8, r Resourcer[T]) *Pool[T] {
	return &Pool[T]{[]*T{}, []*T{}, capacity, r, sync.Mutex{}}
}

func (p *Pool[T]) Get() (*T, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.idle) > 0 {
		resource := p.idle[0]
		p.idle = p.idle[1:]
		p.occupied = append(p.occupied, resource)
		return resource, nil
	}

	if len(p.occupied) < int(p.capacity) {
		resource, err := p.resourcer.Create()
		if err != nil {
			return nil, err
		}
		p.occupied = append(p.occupied, resource)
		return resource, nil
	}

	return nil, errors.New("cannot get more resource")
}

func (p *Pool[T]) Release(resource *T) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := getIdx[T](p.occupied, resource)
	if idx == -1 {
		return errors.New("not found")
	}

	if err := p.resourcer.Reset(); err != nil {
		return err
	}

	p.occupied = append(p.occupied[:idx], p.occupied[idx+1:]...)
	p.idle = append(p.idle, resource)
	return nil
}

func getIdx[T any](slice []*T, target *T) int {
	for idx, v := range slice {
		if target == v {
			return idx
		}
	}
	return -1
}
