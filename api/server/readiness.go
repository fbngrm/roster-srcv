package server

import (
	"net/http"
	"sync/atomic"
)

var healthCode = int32(http.StatusOK)

// HealthCheckShutDown set the health to not ok
func HealthCheckShutDown() {
	atomic.StoreInt32(&healthCode, http.StatusServiceUnavailable)
}

func health() int {
	return int(atomic.LoadInt32(&healthCode))
}

type readinessHandler struct{}

func (h *readinessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(health())
}
