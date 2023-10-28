package pool

import (
	"testing"
	"time"
)

type MockResource struct {
	v int
}

func (r *MockResource) Create() (*MockResource, error) {
	time.Sleep(10 * time.Millisecond)
	return &MockResource{int(time.Now().UnixMicro())}, nil
}

func (r *MockResource) Reset(*MockResource) error {
	return nil
}

const (
	capacity = 5
	overflow = 2
)

func TestNewPool(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Test creating a new pool
	p := NewPool[MockResource](capacity, overflow, r)

	if p.capacity != capacity || p.overflow != overflow || p.currResCnt != capacity {
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
	p := NewPool[MockResource](capacity, overflow, r)

	// Test getting a resource
	res, err := p.Get()
	if err != nil || res == nil {
		t.Errorf("Get did not return a valid resource")
	}

	// Test getting a resource when there are no available idle resources
	// Simulate reaching capacity + overflow
	for i := 0; i < capacity+overflow-1; i++ {
		p.Get()
	}

	res, err = p.Get()
	if err == nil || res != nil {
		t.Errorf("Get did not return an error when no resources are available")
	}
}

func AssertCnt(t *testing.T, cnt uint64) {
	t.Helper()
	if cnt != capacity {
		t.Errorf("Resource count error")
	}
}

func TestRelease(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Create a pool
	p := NewPool[MockResource](capacity, overflow, r)

	// Test releasing a resource
	AssertCnt(t, p.currResCnt)
	res, _ := p.Get()
	AssertCnt(t, p.currResCnt)
	err := p.Release(res)
	AssertCnt(t, p.currResCnt)
	if err != nil {
		t.Errorf("Release returned an error when it shouldn't")
	}

	AssertCnt(t, p.currResCnt)

	// Test releasing a resource when there are available idle resources
	res, _ = p.Get()
	AssertCnt(t, p.currResCnt)
	err = p.Release(res)
	AssertCnt(t, p.currResCnt)
	if err != nil {
		t.Errorf("Release returned an error when it shouldn't")
	}
	AssertCnt(t, p.currResCnt)
}

func TestReleaseWhenIdleFull(t *testing.T) {
	// Create a mock Resourcer
	r := &MockResource{}

	// Create a pool
	p := NewPool[MockResource](capacity, overflow, r)

	res := make([]*MockResource, capacity+overflow)
	for i := 0; i < capacity+overflow; i++ {
		r, _ := p.Get()
		res = append(res, r)
	}

	for i := 0; i < capacity+overflow; i++ {
		p.Release(res[i])
	}

	AssertCnt(t, p.currResCnt)
}
