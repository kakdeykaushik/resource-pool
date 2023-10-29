package pool

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockResource struct{}

func (r *MockResource) Create() (*MockResource, error) {
	time.Sleep(10 * time.Millisecond)
	return &MockResource{}, nil
}

func (r *MockResource) Reset(*MockResource) error {
	return nil
}

const (
	capacity = 5
	waitTime = time.Second
)

// helper
func assertCnt(t *testing.T, cnt uint64) {
	t.Helper()
	if cnt != capacity {
		t.Errorf("Resource count error")
	}
}

// tests - NewPool

func TestNewPool(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Test creating a new pool
	p := NewPool[MockResource](capacity, r)

	if p.capacity != capacity || p.currResCnt != capacity {
		t.Errorf("NewPool did not initialize fields correctly")
	}

	if len(p.idle) != capacity {
		t.Errorf("NewPool did not create the correct number of resources")
	}
}

func TestGet(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Create a pool
	p := NewPool[MockResource](capacity, r)

	// Test getting a resource
	res, err := p.Get()
	if err != nil || res == nil {
		t.Errorf("Get did not return a valid resource")
	}

	// Test getting a resource when there are no available idle resources
	// Simulate reaching capacity + overflow
	for i := 0; i < capacity-1; i++ {
		p.Get()
	}

	res, err = p.Get()
	if err == nil || res != nil {
		t.Errorf("Get did not return an error when no resources are available")
	}

	if err.Error() != "could not get resource" {
		t.Errorf("Incorrect error message: %v", err.Error())
	}
}

func TestRelease(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Create a pool
	p := NewPool[MockResource](capacity, r)

	// Test releasing a resource
	assertCnt(t, p.currResCnt)
	res, _ := p.Get()
	assertCnt(t, p.currResCnt)
	err := p.Release(res)
	assertCnt(t, p.currResCnt)
	if err != nil {
		t.Errorf("Release returned an error when it shouldn't")
	}

	assertCnt(t, p.currResCnt)

	// Test releasing a resource when there are available idle resources
	res, _ = p.Get()
	assertCnt(t, p.currResCnt)
	err = p.Release(res)
	assertCnt(t, p.currResCnt)
	if err != nil {
		t.Errorf("Release returned an error when it shouldn't")
	}
	assertCnt(t, p.currResCnt)
}

func TestReleaseWhenIdleFull(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Create a pool
	p := NewPool[MockResource](capacity, r)

	res := make([]*MockResource, capacity)
	for i := 0; i < capacity; i++ {
		r, _ := p.Get()
		res = append(res, r)
	}

	for i := 0; i < capacity; i++ {
		p.Release(res[i])
	}

	assertCnt(t, p.currResCnt)
}

// tests - NewWaitPool
func TestNewWaitPool(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Test creating a new pool
	p := NewWaitPool[MockResource](capacity, r, waitTime)

	if p.capacity != capacity || p.currResCnt != capacity || p.waitTime != waitTime {
		t.Errorf("NewPool did not initialize fields correctly")
	}

	if len(p.idle) != capacity {
		t.Errorf("NewPool did not create the correct number of resources")
	}
}

func TestWaitedGet(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Create a pool
	p := NewWaitPool[MockResource](capacity, r, waitTime)

	// Test getting a resource
	res, err := p.Get()
	if err != nil || res == nil {
		t.Errorf("Get did not return a valid resource")
	}

	// Test getting a resource when there are no available idle resources
	// Simulate reaching capacity + overflow
	for i := 0; i < capacity-1; i++ {
		p.Get()
	}

	res, err = p.Get()
	if err == nil || res != nil {
		t.Errorf("Get did not return an error when no resources are available")
	}

	if err.Error() != "could not get resource" {
		t.Errorf("Incorrect error message: %v", err.Error())
	}

}

// this is sensitive TC. Bcz [1] should run then, within waitTime [2] must run otherwise TC would fail
// todo: improve this TC
func TestWaitedGetError(t *testing.T) {
	r := &MockResource{}

	// Create a pool
	p := NewWaitPool[MockResource](capacity, r, waitTime)

	// Test getting a resource
	res, err := p.Get()
	if err != nil || res == nil {
		t.Errorf("Get did not return a valid resource")
	}

	for i := 0; i < capacity-1; i++ {
		p.Get()
	}

	go func() {
		p.Release(res) // [2]
	}()

	_, err = p.Get() // [1]
	if err != nil {
		t.Error("e", err.Error())
	}

}

func TestWaitPoolReleaseWhenIdleFull(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Create a pool
	p := NewWaitPool[MockResource](capacity, r, waitTime)

	res := make([]*MockResource, capacity)
	for i := 0; i < capacity; i++ {
		r, _ := p.Get()
		res = append(res, r)
	}

	for i := 0; i < capacity; i++ {
		p.Release(res[i])
	}

	assertCnt(t, p.currResCnt)
}

// tests - Create/Reset errors
type MockErrorResource struct {
	mock.Mock
}

func (r *MockErrorResource) Create() (*MockErrorResource, error) {
	args := r.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MockErrorResource), nil
}

func (r *MockErrorResource) Reset(*MockErrorResource) error {
	return errors.New("error occured while resetting resource")
}

func TestResetError(t *testing.T) {
	r := &MockErrorResource{}

	r.On("Create").Return(nil, errors.New("error occured while creating resource")).Times(capacity)

	p := NewPool[MockErrorResource](capacity, r)

	r.On("Create").Return(&MockErrorResource{}, nil).Once()

	res, err := p.Get()
	if res == nil || err != nil {
		t.Errorf("Should have got resource and nil error")
	}

	err = p.Release(res)
	if err.Error() != "error occured while resetting resource" {
		t.Errorf("Incorrect error message3: %v", err.Error())
	}
}
