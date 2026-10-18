package proxy

import (
	"sync"

	"github.com/stackus/goht/compiler"
)

type SourceMapCache struct {
	sourceMaps map[string]*compiler.SourceMap
	stale      map[string]bool
	mu         sync.Mutex
}

func NewSourceMapCache() *SourceMapCache {
	return &SourceMapCache{
		sourceMaps: make(map[string]*compiler.SourceMap),
		stale:      make(map[string]bool),
	}
}

func (smc *SourceMapCache) Set(uri string, sourceMap *compiler.SourceMap) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	smc.sourceMaps[uri] = sourceMap
	delete(smc.stale, uri)
}

// MarkStale records that the document no longer matches its cached source
// map, such as after an edit that failed to parse.
func (smc *SourceMapCache) MarkStale(uri string) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	smc.stale[uri] = true
}

// IsStale reports whether the cached source map is older than the document.
func (smc *SourceMapCache) IsStale(uri string) bool {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	return smc.stale[uri]
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
	delete(smc.stale, uri)
}
