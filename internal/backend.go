package internal

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	idx   int
	url   *url.URL
	proxy *httputil.ReverseProxy
	mu    sync.RWMutex
	alive bool
}


func (b *Backend) IsAlive() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.alive
}

func (b *Backend) SetAlive(v bool) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.alive = v
}



