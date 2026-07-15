package internal

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	idx   int
	url   *url.URL
	proxy *httputil.ReverseProxy
	mu    sync.RWMutex
	alive bool
	failureCount uint64
	state CircuitState
	openedAt time.Time
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

func(backend *Backend) OnFailure() {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	backend.failureCount++

	switch backend.state {
		case Closed: 
			if backend.failureCount >= 5 {
				backend.state = Open
				backend.openedAt = time.Now()
			}
		case HalfOpen: 
			backend.state = Open
			backend.openedAt = time.Now()	
	}
}

func(backend *Backend) OnSuccess() {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	backend.failureCount = 0

	if backend.state == HalfOpen {
    	backend.state = Closed
	}
}

func(backend *Backend) AllowRequest() bool {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	switch backend.state {
		case Closed: 
			return true
		case Open: 
			if time.Since(backend.openedAt) > 30 * time.Second {
				backend.state = HalfOpen
				return true
			}
		 	return false
		case HalfOpen: 
			return false
	}

	return false
}



