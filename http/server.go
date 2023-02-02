package http

import (
	"encoding/json"
	"github.com/http-nats-proxy/http-proxy/dtos"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"log"
	"net/http"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

type ProxiesHttp interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_, span := otel.Tracer("helloHandler").Start(req.Context(), "Poll")
	defer span.End()
	response, err := dtos.ConvertRequestToQueueRequest(req)
	if err != nil {
		log.Printf("Error converting request to queue data: %v", err)
		http.Error(w, "can't convert request to queue data", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("request_id", response.RequestId))

	err = json.NewEncoder(w).Encode(response.Data)
	if err != nil {
		log.Printf("Error encoding response: %v,%v", err, response.RequestId)
		http.Error(w, "can't encode response", http.StatusBadRequest)
		return

	}
}
