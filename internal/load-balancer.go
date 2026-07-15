package internal

import (
	"load-balancer/metrics"
	"load-balancer/middlewares"
	"load-balancer/request"

	"context"
	"net/http"
	"net/url"
	"sync"
)

type SafeCounter struct {
	mu    sync.Mutex
	count int
}


type LoadBalancer struct {
	servers []*Backend
	counter SafeCounter
	Metrics *metrics.Metrics
}

type RetryState struct {
    Tried map[int]struct{}
}

var retryKey string = "tried_servers"

func (r *RetryState) Try(idx int) bool {
	if _, ok := r.Tried[idx]; ok {
		return false
	}
	r.Tried[idx] = struct{}{}
	return true
}

func (lb *LoadBalancer) AddBackend(rawURL string) error {
    targetURL, err := url.Parse(rawURL)
    if err != nil {
        return err
    }

    lb.servers = append(lb.servers, &Backend{
        url:   targetURL,
        alive: true,
		idx: len(lb.servers),
    })

    lb.servers[len(lb.servers)-1].proxy = NewProxy(lb, lb.servers[len(lb.servers)-1])

	return nil;
}

func (lb* LoadBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	lb.counter.mu.Lock()
	start := lb.counter.count
	
	ctx := req.Context()
	retryState, ok := ctx.Value(retryKey).(*RetryState)

	if !ok {
		retryState = &RetryState{
			Tried: make(map[int]struct{}),
		}

		ctx = context.WithValue(ctx, retryKey, retryState)
		req = req.WithContext(ctx)
	}

	for {
		alive := lb.servers[lb.counter.count].IsAlive()
		allowed := lb.servers[lb.counter.count].AllowRequest()

		if alive && allowed {
			idx := lb.counter.count
			lb.counter.count = (lb.counter.count + 1) % len(lb.servers)
			lb.counter.mu.Unlock()

			ok := retryState.Try(idx)
			
			if !ok {
				http.Error(w, "No healthy backends", http.StatusServiceUnavailable)
				return
			}
			
			lb.servers[idx].proxy.ServeHTTP(w, req)

			return
		}

		lb.counter.count = (lb.counter.count + 1) % len(lb.servers)

		if lb.counter.count == start {
			lb.counter.mu.Unlock()
			http.Error(w, "No healthy backends", http.StatusServiceUnavailable)
			return
		}
	}
}

func (lb *LoadBalancer) ErrorHandler(backend *Backend, w http.ResponseWriter, req *http.Request, err error) {
	if req.Method != http.MethodGet && req.Method != http.MethodOptions && req.Method != http.MethodHead {
		http.Error(w, "Request failed", http.StatusInternalServerError)
		return
	}

	backend.OnFailure()

	start := (backend.idx + 1) % len(lb.servers)
	idx := start
	retryState, ok := req.Context().Value(retryKey).(*RetryState)
	requestState := request.GetRequestState(req)


	if !ok {
		requestState.Recorder.Status = 503
		http.Error(w, "Retry State missing", http.StatusInternalServerError)
		return
	}


	for {
		if lb.servers[idx].IsAlive() && lb.servers[idx].AllowRequest() && retryState.Try(idx)  {
			request.IncrRetry(req)
			lb.servers[idx].proxy.ServeHTTP(w, req)
			return
		}

		idx = (idx + 1)%len(lb.servers)

		if idx == backend.idx {
			break
		}
	}
	
	requestState.Recorder.Status = 503
	http.Error(w, "No healthy backends", http.StatusServiceUnavailable)
}

func(lb *LoadBalancer) Handler() http.Handler {
	return middlewares.ChainMiddlewares(
		lb,
		middlewares.RecorderMiddleware,
		middlewares.RecoveryMiddleware(lb.Metrics),
		middlewares.LoggingMiddleware,
		lb.Metrics.Middleware,
	)
}

func NewLoadBalancer() LoadBalancer {
	return LoadBalancer{ Metrics: &metrics.Metrics{}, }
}

