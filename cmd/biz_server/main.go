package main

import (
	"herostory-server/internal/handler"
	"net/http"
)

func main() {
	http.HandleFunc("/health", handler.HealthCheck)
	http.ListenAndServe(":12345", nil)
}
