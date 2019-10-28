package lrucache

import (
	"fmt"
	"testing"
)

const (
	testCapacity = 10
	testIter     = 100
)

func TestCache(t *testing.T) {
	cache := NewCache(testCapacity)

	// insert test
	for i := 0; i < testIter; i++ {
		cache.Set(fmt.Sprintf("%d", i), i)
	}

	if cache.evicted != testIter-testCapacity {
		t.Errorf("cache evicted count is not match (%d != %d)", cache.evicted, testIter-testCapacity)
		t.FailNow()
	}

	// get test
	i := (testCapacity / 2) + (testIter - testCapacity)
	key := fmt.Sprintf("%d", i)

	retrieved, ok, _ := cache.Get(key)
	if !ok {
		t.Errorf("can't found: %v", key)
		t.FailNow()
	}

	retI, ok := retrieved.(int)
	if !ok {
		t.Errorf("invalid data type")
		t.FailNow()
	}

	if retI != i {
		t.Errorf("retrieved i is not matched (%d != %d)", retI, i)
		t.FailNow()
	}

	headI := cache.head.data.(int)

	if headI != i {
		t.Errorf("head i is not matched (%d != %d)", headI, i)
		t.FailNow()
	}

	curr := cache.head
	count := 0

	for curr != nil {
		count++
		curr = curr.next
	}

	if len(cache.m) != testCapacity {
		t.Errorf("stored number of items(in map) are not matched (%d != %d)", len(cache.m), testCapacity)
		t.FailNow()
	}

	if count != testCapacity {
		t.Errorf("stored number of items(in list) are not matched (%d != %d)", count, testCapacity)
		t.FailNow()
	}
}
