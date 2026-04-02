package cache

import (
	"sync"

	"github.com/dgraph-io/ristretto"
)

// RistrettoCache Ristretto内存缓存封装
type RistrettoCache struct {
	cache *ristretto.Cache
	lock  *sync.RWMutex
}

func NewRistrettoCache() (*RistrettoCache, error) {
	c := &ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30, // 1GB
		BufferItems: 64,
	}

	cache, err := ristretto.NewCache(c)
	if err != nil {
		return nil, err
	}

	return &RistrettoCache{
		cache: cache,
		lock:  new(sync.RWMutex),
	}, nil
}

// Get 获取缓存
func (r *RistrettoCache) Get(key string) (interface{}, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.cache.Get(key)
}

// Set 设置缓存
func (r *RistrettoCache) Set(key string, value interface{}, cost int64) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.cache.Set(key, value, cost)
}

// Clear 清空缓存
func (r *RistrettoCache) Clear() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.cache.Clear()
}

// Delete 删除指定键的缓存
func (r *RistrettoCache) Delete(key string) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.cache.Del(key)
	return true
}
