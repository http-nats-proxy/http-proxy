package main

import (
	h "github.com/http-nats-proxy/http-proxy/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log"
	"net/http"
)

func main() {
	defer initLogging(0)()
	// Hello world, the web server
	server := h.NewDelayedEchoServer(250)
	Info("Listing for requests", "url", "http://localhost:8000/echo")
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(server.ServeHTTP), "Proxy")
	http.Handle("/", otelHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))

}
