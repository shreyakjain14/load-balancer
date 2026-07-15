package middlewares

import (
	"context"
	"load-balancer/request"

	"net/http"
)

func RecorderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &request.StatusRecorder{
			ResponseWriter: w,
			Status: http.StatusOK,
		}

		state := &request.RequestState{
			Recorder: rec,
		}

		ctx := context.WithValue(r.Context(), request.RecordStateKey{}, state)

		next.ServeHTTP(rec, r.WithContext(ctx))
	})
}

