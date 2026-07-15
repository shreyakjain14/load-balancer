package middlewares

import (
	"load-balancer/request"
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Started %s %s", req.Method, req.URL.Path)
		start := time.Now()

		defer func() {
    		elapsed := time.Since(start)
			reqState := request.GetRequestState(req)

    		log.Printf(
        		"%s %s status=%v duration=%v",
       			req.Method,
        		req.URL.Path,
        		reqState.Recorder.Status,
        		elapsed,
   		 	)
		}()

		next.ServeHTTP(w, req)
	})
}
