package main

import (
	"fmt"
	"sync"
)

type FilterCache struct {
	mu sync.RWMutex
	cache map[string] *AppVer
}

func NewFilterCache () *FilterCache {
	return &FilterCache{cache:make(map[string] *AppVer)}
}

func (fc *FilterCache) Get(key string) (value *AppVer, ok bool) {
	fc.mu.RLock()
	value, ok = fc.cache[key]
	fc.mu.RUnlock()
	return
}

func (fc *FilterCache) set(key string, value *AppVer) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.cache[key] = value
}

func (fc *FilterCache) toString () {
	fmt.Sprintf("%v", fc)
}