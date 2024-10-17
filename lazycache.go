package lazycache

// A generic cache that favors returning stale data
// than blocking a caller

import (
	"sync"
	"time"
)

type Fetcher func(id string) (interface{}, error)
type GroupFetcher func() (map[string]interface{}, error)

type Item struct {
	object  interface{}
	expires time.Time
}

type LazyCache struct {
	fetcher      Fetcher
	groupFetcher GroupFetcher
	ttl          time.Duration
	lock         sync.RWMutex
	items        map[string]*Item
}

func New(fetcher Fetcher, ttl time.Duration, size int) *LazyCache {
	return &LazyCache{
		ttl:          ttl,
		fetcher:      fetcher,
		groupFetcher: nil,
		items:        make(map[string]*Item, size),
	}
}

func NewGroup(groupFetcher GroupFetcher, fetcher Fetcher, ttl time.Duration, size int) *LazyCache {
	return &LazyCache{
		ttl:          ttl,
		fetcher:      fetcher,
		groupFetcher: groupFetcher,
		items:        make(map[string]*Item, size),
	}
}

func (cache *LazyCache) SwapCache(fetcher Fetcher, groupFetcher GroupFetcher) *LazyCache {
	cache.fetcher = fetcher
	cache.groupFetcher = groupFetcher
	return cache
}

func (cache *LazyCache) Get(id string) (interface{}, bool) {
	cache.lock.RLock()
	item, exists := cache.items[id]
	if !exists {
		cache.lock.RUnlock()
		if cache.groupFetcher != nil {
			return cache.groupFetch(id)
		}
		return cache.Fetch(id)
	}
	expires := item.expires
	object := item.object
	cache.lock.RUnlock()
	if time.Now().After(expires) {
		if cache.groupFetcher != nil {
			go cache.groupFetch(id)
		} else {
			go cache.Fetch(id)
		}
	}
	return object, object != nil
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

func (cache *LazyCache) Fetch(id string) (interface{}, bool) {
	object, err := cache.fetcher(id)
	if err != nil {
		return nil, false
	}
	cache.Set(id, object)
	return object, object != nil
}

func (cache *LazyCache) groupFetch(id string) (interface{}, bool) {
	objects, err := cache.groupFetcher()
	if err != nil || objects == nil {
		if cache.fetcher == nil {
			return nil, false
		}
		return cache.Fetch(id) // fallback to single fetch
	}

	var res interface{}
	for k, v := range objects {
		cache.Set(k, v)
		if k == id {
			res = v
		}
	}
	return res, res != nil
}
