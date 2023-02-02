package http

import (
	"encoding/json"
	"github.com/http-nats-proxy/http-proxy/dtos"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"time"
)

type DelayedEchoServer struct {
	DelayDuration int64
	Tracer        trace.Tracer
}

func NewDelayedEchoServer(delayDuration int64) *DelayedEchoServer {
	tracer := otel.Tracer("DelayedEchoServer")

	return &DelayedEchoServer{
		DelayDuration: delayDuration,
		Tracer:        tracer,
	}
}

func (s DelayedEchoServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_, span := s.Tracer.Start(req.Context(), "ServeHTTP")
	defer span.End()
	response, err := dtos.ConvertRequestToQueueRequest(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Printf("Error converting request to queue data: %v", err)
		http.Error(w, "can't convert request to queue data", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.String("requestId", response.RequestId))

	span.AddEvent("Sending request to queue", trace.WithTimestamp(time.Now()))

	startTime := time.Now()
	time.Sleep(time.Duration(s.DelayDuration) * time.Millisecond)
	endTime := time.Now()

	span.AddEvent("got response from queue",
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(attribute.Int64("duration", endTime.Sub(startTime).Milliseconds())),
	)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Printf("Error encoding response: %v,%v", err, response.RequestId)
		http.Error(w, "can't encode response", http.StatusBadRequest)
		return

	}
	span.SetStatus(codes.Ok, "")
}
