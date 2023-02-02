package main

import (
	h "github.com/http-nats-proxy/http-proxy/http"
	"log"
	"net/http"
)

func main() {
	defer initLogging(0)()
	// Hello world, the web server
	server := h.NewDelayedEchoServer(250)
	http.HandleFunc("/", server.ServeHTTP)
	Info("Listing for requests", "url", "http://localhost:8000/echo")
	log.Fatal(http.ListenAndServe(":8000", nil))

}
