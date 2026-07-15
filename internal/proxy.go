package internal

import (
	"net/http"
	"net/http/httputil"
)


func NewProxy(lb *LoadBalancer, backend *Backend) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(backend.url)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		lb.ErrorHandler(backend, w, r, err)
	}

	return proxy
}