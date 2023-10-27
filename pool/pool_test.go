package pool

import (
	"testing"
)

type MockResource struct{}

func (m *MockResource) Create() (*MockResource, error) { return &MockResource{}, nil }
func (m *MockResource) Reset() error                   { return nil }

func TestGet(t *testing.T) {
	mockResource := &MockResource{}
	p := NewPool[MockResource](1, mockResource)
	// Test case: Getting a resource when there are available resources
	resource, err := p.Get()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resource == nil {
		t.Error("Expected a resource, got nil")
	}

	// Test case: Trying to get more resources than the capacity
	_, err = p.Get()
	if err == nil || err.Error() != "cannot get more resource" {
		t.Errorf("Expected 'cannot get more resource' error, got %v", err)
	}
}

func TestRelease(t *testing.T) {
	mockResource := &MockResource{}
	p := NewPool[MockResource](2, mockResource)

	// Test case: Releasing a resource that is in the occupied list
	resource, _ := p.Get()
	err := p.Release(resource)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestReleaseNotFound(t *testing.T) {
	mockResource := &MockResource{}
	p := NewPool[MockResource](2, mockResource)

	// Test case: Releasing a resource that is not in the occupied list
	err := p.Release(&MockResource{})
	if err == nil || err.Error() != "not found" {
		t.Errorf("Expected 'not found' error, got %v", err)
	}
}

func TestReleaseAndGet(t *testing.T) {
	mockResource := &MockResource{}
	p := NewPool[MockResource](1, mockResource)
	// Test case: Getting a resource when there are available resources
	resource, err := p.Get()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resource == nil {
		t.Error("Expected a resource, got nil")
	}

	p.Release(resource)
	resource, err = p.Get()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resource == nil {
		t.Error("Expected a resource, got nil")
	}
}
