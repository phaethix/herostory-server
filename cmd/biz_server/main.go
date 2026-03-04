package main

import (
	"herostory-server/internal/bootstrap"
	"herostory-server/internal/server"
	"net/http"
)

func main() {
	bootstrap.InitApp()

	http.HandleFunc("/health", server.HealthCheck)
	http.HandleFunc("/websocket", server.WebSocketHandshake)
	http.ListenAndServe(":12345", nil)
}
