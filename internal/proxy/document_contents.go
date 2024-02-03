package proxy

import (
	"fmt"
	"sync"

	"github.com/stackus/goht/internal/protocol"
)

type DocumentContents struct {
	documents map[string]*Document
	mu        sync.Mutex
}

func NewDocumentContents() *DocumentContents {
	return &DocumentContents{
		documents: make(map[string]*Document),
	}
}

func (dc *DocumentContents) Set(uri string, d *Document) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.documents[uri] = d
}

func (dc *DocumentContents) Get(uri string) (d *Document, ok bool) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	d, ok = dc.documents[uri]
	return
}

func (dc *DocumentContents) Delete(uri string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	delete(dc.documents, uri)
}

func (dc *DocumentContents) URIs() (uris []string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	uris = make([]string, len(dc.documents))
	var i int
	for k := range dc.documents {
		uris[i] = k
		i++
	}
	return uris
}

func (dc *DocumentContents) Apply(uri string, changes []protocol.TextDocumentContentChangeEvent) (*Document, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	d, ok := dc.documents[uri]
	if !ok {
		return nil, fmt.Errorf("document %q not found", uri)
	}

	for _, change := range changes {
		d.Apply(change.Range, change.Text)
	}

	return d, nil
}
