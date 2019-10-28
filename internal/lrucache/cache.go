package lrucache

import (
	"errors"
	"sync"
)

type Cache struct {
	capacity int

	m map[string]*entry

	head *entry
	tail *entry

	evicted int

	lock *sync.Mutex
}

type entry struct {
	key  string
	data interface{}

	prev *entry
	next *entry
}

func NewCache(capacity int) *Cache {
	return &Cache{
		capacity: capacity,
		m:        make(map[string]*entry),
		head:     nil,
		tail:     nil,
		evicted:  0,
		lock:     new(sync.Mutex),
	}
}

func (cache *Cache) recentlyUsed(e *entry) {
	head := cache.head

	switch head {
	case nil:
		panic("head must be existed")
	case e:
		// there are only one element
		return
	}

	// disconnect element
	prev := e.prev
	next := e.next

	if prev != nil {
		prev.next = next
	}

	if next != nil {
		next.prev = prev
	}

	// prepend to head
	head.prev = e
	e.prev = nil
	e.next = head
	cache.head = e

	return
}

func (cache *Cache) Get(key string) (interface{}, bool, error) {
	if key == "" {
		return nil, false, nil
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()

	e := cache.m[key]
	if e == nil {
		return nil, false, nil
	}

	cache.recentlyUsed(e)
	return e.data, true, nil
}

func (cache *Cache) Set(key string, data interface{}) error {
	if key == "" {
		return errors.New("empty key")
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()

	e := cache.m[key]
	if e != nil {
		// just update entry value
		e.data = data
		cache.recentlyUsed(e)

		return nil
	}

	// check capacity
	if len(cache.m) >= cache.capacity {
		// evict least recently used entry (i.e., tail)
		tail := cache.tail
		if tail == nil {
			panic("tail must be not nil")
		}

		tailPrev := tail.prev

		if tailPrev != nil {
			// disconnect tail
			tailPrev.next = nil
		}

		delete(cache.m, tail.key)
		cache.tail = tailPrev

		cache.evicted++
	}

	// make a new entry
	newE := &entry{
		key:  key,
		data: data,
	}

	cache.m[key] = newE

	oldHead := cache.head

	if oldHead == nil { // first element
		if cache.tail != nil {
			panic("if head is nil, tail must be nil")
		}

		cache.head = newE
		cache.tail = newE

		return nil
	}

	oldHead.prev = newE
	newE.next = oldHead
	cache.head = newE

	return nil
}
