package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Resourcer[T any] interface {
	Create() (*T, error)
	Reset(*T) error
}

type Pool[T any] struct {
	idle       chan *T
	capacity   uint32
	currResCnt uint64
	resourcer  Resourcer[T]
	// mu         sync.Mutex
	waitTime time.Duration
}

func spinRes[T any](capacity uint32, r Resourcer[T]) (chan *T, uint64) {
	resources := make(chan *T, capacity)
	var resCnt uint64 = 0
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
	return resources, resCnt
}

func NewPool[T any](capacity uint32, r Resourcer[T]) *Pool[T] {
	resources, resCnt := spinRes[T](capacity, r)
	return &Pool[T]{resources, capacity, resCnt, r, 0}
}

func NewWaitPool[T any](capacity uint32, r Resourcer[T], waitTime time.Duration) *Pool[T] {
	resources, resCnt := spinRes[T](capacity, r)
	return &Pool[T]{resources, capacity, resCnt, r, waitTime}
}

func (p *Pool[T]) isFull() bool {
	return atomic.LoadUint64(&p.currResCnt) >= uint64(p.capacity)
}

func (p *Pool[T]) Get() (*T, error) {

	// edge case
	if len(p.idle) == 0 && (!p.isFull()) {
		res, err := p.resourcer.Create()
		if err == nil && res != nil {
			atomic.AddUint64(&p.currResCnt, 1)
			p.idle <- res
		}
	}

	if len(p.idle) == 0 {
		select {
		case <-time.After(p.waitTime):
			return nil, errors.New("could not get resource")
		case r := <-p.idle:
			return r, nil
		}
	}

	// possibility of bug here ?
	// for a non-waiting Pool, if idle channel has 1 resource and
	// 2 concurrent instructions reaches this line of code then 1 of them have
	// to wait until resource is being put back into Pool
	return <-p.idle, nil
}

func (p *Pool[T]) Release(resource *T) error {

	err := p.resourcer.Reset(resource)
	if err != nil {
		return err
	}

	p.idle <- resource
	return nil
}
