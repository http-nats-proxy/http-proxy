package main

import (
	h "github.com/http-nats-proxy/http-proxy/http"
	"log"
	"net/http"
)

func main() {
	// Hello world, the web server
	server := h.NewEchoServer()
	http.HandleFunc("/", server.ServeHTTP)
	log.Println("Listing for requests at http://localhost:8000/echo")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
