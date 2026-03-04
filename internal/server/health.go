package server

import "net/http"

func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("healthy"))
}
