package main

import (
	"github.com/http-nats-proxy/http-proxy/global_logger"
	h "github.com/http-nats-proxy/http-proxy/http"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log"
	"net/http"
)

func main() {
	app := global_logger.InitLogging(9)
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("Error shutting down: %v", err)
		}
	}()
	topic := "topic"

	natsUrl := nats.DefaultURL
	app.Logger.Debug("Connection to nats", "url", natsUrl)
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		app.Logger.Error(err, "failed to connect to nats", "url", natsUrl)
	}

	server := h.NewNatsPublishServer(app, nc, h.NatsRequester{Conn: nc}, topic)
	// Hello world, the web server
	app.Logger.Info("Listing for requests", "url", "http://localhost:8000/echo")
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(server.ServeHTTP), "Proxy")
	http.Handle("/", otelHandler)

	err = http.ListenAndServe(":8000", nil)
	nc.Close()
	log.Fatal(err)

}
