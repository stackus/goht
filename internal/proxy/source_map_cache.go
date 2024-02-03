package proxy

import (
	"sync"

	"github.com/stackus/goht/compiler"
)

type SourceMapCache struct {
	sourceMaps map[string]*compiler.SourceMap
	mu         sync.Mutex
}

func NewSourceMapCache() *SourceMapCache {
	return &SourceMapCache{
		sourceMaps: make(map[string]*compiler.SourceMap),
	}
}

func (smc *SourceMapCache) Set(uri string, sourceMap *compiler.SourceMap) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	smc.sourceMaps[uri] = sourceMap
}

func (smc *SourceMapCache) Get(uri string) (sourceMap *compiler.SourceMap, ok bool) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	sourceMap, ok = smc.sourceMaps[uri]
	return
}

func (smc *SourceMapCache) Delete(uri string) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	delete(smc.sourceMaps, uri)
}
