package request

import "net/http"

type RetryState struct {
	Count uint64
}

type StatusRecorder struct {
	http.ResponseWriter
	HasWritten bool
	Status int
	Retry RetryState
}

type RequestState struct {
	Recorder *StatusRecorder
}

type RecordStateKey struct{}

func (r *StatusRecorder) WriteHeader(code int) {
	if r.HasWritten {
		return
	}
	r.HasWritten = true
	r.Status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *StatusRecorder) Write(b []byte) (int, error) {
	if !r.HasWritten {
		r.Status = http.StatusOK
		r.HasWritten = true
	}

	return r.ResponseWriter.Write(b)
}

func GetRequestState(r *http.Request) *RequestState {
    state, ok := r.Context().Value(RecordStateKey{}).(*RequestState)
    if !ok {
        panic("request state missing")
    }
    return state
}

func IncrRetry(req *http.Request) {
	state := GetRequestState(req)
	state.Recorder.Retry.Count++	
}