package middlewares

import (
	"load-balancer/metrics"
	"load-balancer/request"
	"log"
	"net/http"
)

func RecoveryMiddleware(metrics *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				requestState := request.GetRequestState(req)

				if r:= recover(); r != nil {
					metrics.IncrPanics()
					requestState.Recorder.Status = 500
					log.Printf("Recovered. Error%v\n", r)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, req)
		})
	}
}