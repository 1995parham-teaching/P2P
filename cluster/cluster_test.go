package cluster

import (
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	list := []string{"127.0.0.1:1378", "192.168.1.1:1379"}
	c := New(list)

	if c == nil {
		t.Fatal("New() returned nil")
	}

	result := c.List()
	if len(result) != len(list) {
		t.Errorf("List() length = %d, want %d", len(result), len(list))
	}

	// Verify it's a copy and not the original
	list[0] = "modified"
	if c.List()[0] == "modified" {
		t.Error("New() should copy the list, not use the original reference")
	}
}

func TestListReturnsACopy(t *testing.T) {
	c := New([]string{"127.0.0.1:1378"})

	list1 := c.List()
	list2 := c.List()

	// Modify one list
	list1[0] = "modified"

	// The other should be unchanged
	if list2[0] == "modified" {
		t.Error("List() should return a copy, not the internal slice")
	}

	// Internal state should be unchanged
	if c.List()[0] == "modified" {
		t.Error("List() should return a copy, modifications should not affect internal state")
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name         string
		initial      []string
		host         string
		toMerge      []string
		expectedSize int
	}{
		{
			name:         "merge new addresses",
			initial:      []string{"127.0.0.1:1378"},
			host:         "127.0.0.1:1000",
			toMerge:      []string{"192.168.1.1:1379", "10.0.0.1:1380"},
			expectedSize: 3,
		},
		{
			name:         "skip duplicates",
			initial:      []string{"127.0.0.1:1378"},
			host:         "127.0.0.1:1000",
			toMerge:      []string{"127.0.0.1:1378", "192.168.1.1:1379"},
			expectedSize: 2,
		},
		{
			name:         "skip host address",
			initial:      []string{"127.0.0.1:1378"},
			host:         "127.0.0.1:1000",
			toMerge:      []string{"127.0.0.1:1000", "192.168.1.1:1379"},
			expectedSize: 2,
		},
		{
			name:         "skip empty strings",
			initial:      []string{"127.0.0.1:1378"},
			host:         "127.0.0.1:1000",
			toMerge:      []string{"", "192.168.1.1:1379"},
			expectedSize: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.initial)
			c.Merge(tt.host, tt.toMerge)

			if c.Size() != tt.expectedSize {
				t.Errorf("Size() = %d, want %d", c.Size(), tt.expectedSize)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	c := New([]string{})

	c.Add("127.0.0.1:1378")
	if c.Size() != 1 {
		t.Errorf("Size() = %d, want %d", c.Size(), 1)
	}

	// Adding duplicate should not increase size
	c.Add("127.0.0.1:1378")
	if c.Size() != 1 {
		t.Errorf("Size() after duplicate = %d, want %d", c.Size(), 1)
	}

	// Adding empty string should not increase size
	c.Add("")
	if c.Size() != 1 {
		t.Errorf("Size() after empty = %d, want %d", c.Size(), 1)
	}

	// Adding new address should increase size
	c.Add("192.168.1.1:1379")
	if c.Size() != 2 {
		t.Errorf("Size() after new address = %d, want %d", c.Size(), 2)
	}
}

func TestRemove(t *testing.T) {
	c := New([]string{"127.0.0.1:1378", "192.168.1.1:1379"})

	c.Remove("127.0.0.1:1378")
	if c.Size() != 1 {
		t.Errorf("Size() = %d, want %d", c.Size(), 1)
	}

	// Removing non-existent should be safe
	c.Remove("10.0.0.1:1380")
	if c.Size() != 1 {
		t.Errorf("Size() after removing non-existent = %d, want %d", c.Size(), 1)
	}

	c.Remove("192.168.1.1:1379")
	if c.Size() != 0 {
		t.Errorf("Size() after removing all = %d, want %d", c.Size(), 0)
	}
}

func TestSize(t *testing.T) {
	c := New([]string{})
	if c.Size() != 0 {
		t.Errorf("Size() = %d, want %d", c.Size(), 0)
	}

	c = New([]string{"127.0.0.1:1378", "192.168.1.1:1379"})
	if c.Size() != 2 {
		t.Errorf("Size() = %d, want %d", c.Size(), 2)
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := New([]string{"127.0.0.1:1378"})

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent adds
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Add("10.0.0.1:" + string(rune(i+1000)))
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.List()
			_ = c.Size()
		}()
	}

	// Concurrent merges
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Merge("127.0.0.1:1000", []string{"192.168.1.1:1379"})
		}()
	}

	wg.Wait()

	// Should not panic and list should be accessible
	list := c.List()
	if len(list) == 0 {
		t.Error("Expected non-empty list after concurrent operations")
	}
}
