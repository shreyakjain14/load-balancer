package metrics

import (
	"encoding/json"
	"load-balancer/request"
	"net/http"
	"sync/atomic"
)

type Metrics struct {
	TotalRequests      uint64
	SuccessfulRequests uint64
	FailedRequests     uint64
	RetryCount         uint64
	PanicCount         uint64
}

type MetricsResponse struct {
	TotalRequests      uint64 `json:"totalRequests"`
	SuccessfulRequests uint64 `json:"successfulRequests"`
	FailedRequests     uint64 `json:"failedRequests"`
	RetryCount         uint64 `json:"retryCount"`
	PanicCount         uint64 `json:"panicCount"`
}

func (m *Metrics) IncrRequests() {
	atomic.AddUint64(&m.TotalRequests, 1)
}

func(m *Metrics) IncrSuccessfulRequests(){
	atomic.AddUint64(&m.SuccessfulRequests,1)
}

func (m *Metrics) IncrFailedRequests() {
    atomic.AddUint64(&m.FailedRequests, 1)
}

func (m *Metrics) IncrRetriesBy(num uint64) {
    atomic.AddUint64(&m.RetryCount, num)
}

func (m *Metrics) IncrPanics() {
    atomic.AddUint64(&m.PanicCount, 1)
}

func(m *Metrics) LoadRequests() uint64 {
	return atomic.LoadUint64(&m.TotalRequests)
}


func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc( func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			requestState := request.GetRequestState(r)
			m.IncrRetriesBy(requestState.Recorder.Retry.Count)

			switch {
				case requestState.Recorder.Status >= 200 && requestState.Recorder.Status < 400:
   	 				m.IncrSuccessfulRequests()
				default:
    				m.IncrFailedRequests()
			}
			
		}()

		m.IncrRequests()
		next.ServeHTTP(w, r)
	})
}

func(m *Metrics) Snapshot() MetricsResponse {
	return MetricsResponse{
        TotalRequests:      atomic.LoadUint64(&m.TotalRequests),
        SuccessfulRequests: atomic.LoadUint64(&m.SuccessfulRequests),
        FailedRequests:     atomic.LoadUint64(&m.FailedRequests),
        RetryCount:         atomic.LoadUint64(&m.RetryCount),
        PanicCount:         atomic.LoadUint64(&m.PanicCount),
    }
}

func(m *Metrics) Handler(w http.ResponseWriter, r *http.Request) {
	snapshot := m.Snapshot()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}