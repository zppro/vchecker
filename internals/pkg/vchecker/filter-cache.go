package vchecker

import (
	"fmt"
	"sync"
	"github.com/zppro/vchecker/internals/pkg/shared"
)

type FilterCache struct {
	mu sync.RWMutex
	cache map[string] *shared.AppVer
}

func NewFilterCache () *FilterCache {
	return &FilterCache{cache:make(map[string] *shared.AppVer)}
}

func (fc *FilterCache) Get(key string) (value *shared.AppVer, ok bool) {
	fc.mu.RLock()
	value, ok = fc.cache[key]
	fc.mu.RUnlock()
	return
}

func (fc *FilterCache) Set(key string, value *shared.AppVer) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.cache[key] = value
}

func (fc *FilterCache) Clear() (value *shared.AppVer, ok bool) {
	fc.mu.RLock()
	fc.cache = make(map[string] *shared.AppVer)
	fc.mu.RUnlock()
	return
}

func (fc *FilterCache) toString () {
	fmt.Sprintf("%v", fc)
}