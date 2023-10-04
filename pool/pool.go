package pool

import (
	"errors"
)

// todo: test resource
type Pool[T any] struct {
	idle     []*T
	occupied []*T
	capacity uint8
	create   func() (*T, error) // todo: better
}

func NewPool[T any](capacity uint8, create func() (*T, error)) *Pool[T] {
	return &Pool[T]{[]*T{}, []*T{}, capacity, create}
}

func (p *Pool[T]) Get() (*T, error) {

	if len(p.idle) > 0 {
		resource := p.idle[0]
		p.idle = p.idle[1:]
		p.occupied = append(p.occupied, resource)
		return resource, nil
	}

	if len(p.occupied) < int(p.capacity) {
		resource, err := p.create()
		if err != nil {
			return nil, err
		}
		p.occupied = append(p.occupied, resource)
		return resource, nil
	}

	return nil, errors.New("can get more resource")
}

func (p *Pool[T]) Release(resource *T) error {
	idx := getIdx[T](p.occupied, resource)
	if idx == -1 {
		return errors.New("not found")
	}

	p.occupied = append(p.occupied[:idx], p.occupied[idx+1:]...)
	p.idle = append(p.idle, resource)
	return nil
}

func getIdx[T any](slice []*T, target *T) int {
	for idx, v := range slice {
		// todo: check this
		if target == v {
			return idx
		}
	}
	return -1
}
