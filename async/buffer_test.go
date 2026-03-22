package async

import (
	"sync"
	"testing"
)

func TestBufferQueue_PushPop(t *testing.T) {
	q := NewBufferQueue[int]()

	// Initial pop on empty
	res := q.PopAll()
	if len(res) != 0 {
		t.Errorf("Expected empty buffer to return nil, got len %d", len(res))
	}

	q.Push(1)
	q.Push(2)
	q.Push(3)

	res = q.PopAll()
	if len(res) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(res))
	}

	expected := []int{1, 2, 3}
	for i, v := range res {
		if v != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
		}
	}

	res2 := q.PopAll()
	if len(res2) != 0 {
		t.Errorf("Expected empty buffer after PopAll, got len %d", len(res2))
	}
}

func TestBufferQueue_Concurrency(t *testing.T) {
	q := NewBufferQueue[int]()
	var wg sync.WaitGroup

	// 100 goroutines pushing 100 items each
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				q.Push(id*100 + j)
			}
		}(i)
	}

	wg.Wait()

	res := q.PopAll()
	if len(res) != 10000 {
		t.Errorf("Expected 10000 items, got %d", len(res))
	}

	// Verify capacity is preserved
	if cap(q.active) < 10000 {
		t.Errorf("Expected active capacity to be preserved, got %d", cap(q.active))
	}
}

func TestBufferQueue_MemoryLeak(t *testing.T) {
	q := NewBufferQueue[*int]()

	val1 := 5
	val2 := 10

	q.Push(&val1)
	q.Push(&val2)
	q.PopAll() // This should set internal array slots to nil

	// Because we can't directly check the internal array using safe Go code
	// without reaching into internals, we inspect q.active's backing array capacity
	// to see that it isn't completely nuked, but the elements inside its original length are nil
	
	// We pushed 2 elements, so cap should be at least 2.
	// Since length is 0, we can temporarily slice it back to 2 to see the backing array.
	backingArray := q.active[:2]

	for i, p := range backingArray {
		if p != nil {
			t.Errorf("Memory leak detected: element %d in backing array is not nil", i)
		}
	}
}
