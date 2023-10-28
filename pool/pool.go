package pool

import (
	"errors"
	"sync"
	"sync/atomic"
)

type Resourcer[T any] interface {
	Create() (*T, error)
	Reset(*T) error
}

type Pool[T any] struct {
	idle       chan *T
	capacity   uint32
	overflow   uint32
	currResCnt uint64
	resourcer  Resourcer[T]
	mu         sync.Mutex
}

func NewPool[T any](capacity uint32, overflow uint32, r Resourcer[T]) *Pool[T] {
	resources := make(chan *T, capacity+overflow)
	resCnt := uint64(0)
	var wg sync.WaitGroup

	wg.Add(int(capacity))
	for i := 0; i < int(capacity); i++ {
		go func() {
			defer wg.Done()
			res, err := r.Create()
			if err != nil {
				return
			}
			atomic.AddUint64(&resCnt, 1)
			resources <- res
		}()
	}

	wg.Wait()
	return &Pool[T]{resources, capacity, overflow, resCnt, r, sync.Mutex{}}
}

func (p *Pool[T]) Get() (*T, error) {

	if len(p.idle) == 0 && atomic.LoadUint64(&p.currResCnt) < uint64(p.capacity)+uint64(p.overflow) {
		res, err := p.resourcer.Create()
		if err != nil {
			return nil, errors.New("could not get resource")
		}
		atomic.AddUint64(&p.currResCnt, 1)
		p.idle <- res
	}

	switch len(p.idle) {
	case 0:
		return nil, errors.New("could not get resource")
	default:
		r := <-p.idle
		return r, nil
	}
}

func (p *Pool[T]) Release(resource *T) error {

	err := p.resourcer.Reset(resource)
	if err != nil {
		return err
	}

	if len(p.idle) < int(p.capacity) {
		p.idle <- resource
		return nil
	}

	atomic.AddUint64(&p.currResCnt, ^uint64(0))
	// cannot put it in pool, so resource should be destroyed
	// Go being GC language we cannot delete the object, GC will take care of it
	// todo: re check this, you might have learnt something in time being
	return nil
}
