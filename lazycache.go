package lazycache

// A generic cache that favors returning stale data
// than blocking a caller

import (
	"sync"
	"time"
)

type Fetcher func(id string) (interface{}, error)

type Item struct {
	object  interface{}
	expires time.Time
}

type LazyCache struct {
	fetcher Fetcher
	ttl     time.Duration
	lock    sync.RWMutex
	items   map[string]*Item
}

func New(fetcher Fetcher, ttl time.Duration, size int) *LazyCache {
	return &LazyCache{
		ttl:     ttl,
		fetcher: fetcher,
		items:   make(map[string]*Item, size),
	}
}

func (cache *LazyCache) Get(id string) (interface{}, bool, error) {
	cache.lock.RLock()
	item, exists := cache.items[id]
	if exists == false {
		cache.lock.RUnlock()
		return cache.Fetch(id)
	}
	expires := item.expires
	object := item.object
	cache.lock.RUnlock()
	if time.Now().After(expires) {
		go cache.Fetch(id)
	}
	return object, object != nil, nil
}

func (cache *LazyCache) Set(id string, object interface{}) {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	current, exists := cache.items[id]
	if exists {
		current.expires = time.Now().Add(cache.ttl)
		current.object = object
	} else {
		cache.items[id] = &Item{expires: time.Now().Add(cache.ttl), object: object}
	}
}

func (cache *LazyCache) Fetch(id string) (interface{}, bool, error) {
	object, err := cache.fetcher(id)
	if err != nil {
		return nil, false, err
	}
	cache.Set(id, object)
	return object, object != nil, nil
}
