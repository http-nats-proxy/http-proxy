package http

import (
	"context"
	"encoding/json"
	"github.com/http-nats-proxy/http-proxy/dtos"
	. "github.com/http-nats-proxy/http-proxy/global_logger"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"time"
)

type NatsRequester struct {
	*nats.Conn
}

func (r NatsRequester) Request(ctx context.Context, span trace.Span, data []byte, topic string) (headers *http.Header, bytes []byte, err error) {
	span.AddEvent("RequestWithContext Start")
	msg := nats.NewMsg(topic)
	msg.Data = data
	sc := span.SpanContext()
	trace.SpanFromContext(ctx)
	if sc.IsValid() {
		spanBytes, err := span.SpanContext().MarshalJSON()
		if err != nil {
			span.AddEvent("failed to marshal context ", trace.WithAttributes(attribute.String("error", err.Error())))
			span.RecordError(err)
			return nil, nil, err
		}
		log.Println(string(spanBytes))
		msg.Header.Set("TraceID", sc.TraceID().String())
		msg.Header.Set("TraceFlags", sc.TraceFlags().String())
		msg.Header.Set("TraceState", sc.TraceState().String())
		msg.Header.Set("SpanID", sc.SpanID().String())
	}

	responseMessage, err := r.RequestMsgWithContext(ctx, msg)
	if err != nil {
		span.AddEvent("RequestWithContext Failed")
		span.RecordError(err)
		return nil, nil, err
	}
	span.AddEvent("RequestWithContext OK")
	h := http.Header(responseMessage.Header)

	return &h, responseMessage.Data, nil
}

type TopicRequester interface {
	Request(ctx context.Context, span trace.Span, data []byte, topic string) (headers *http.Header, bytes []byte, err error)
}

type NatsPublishServer struct {
	App       *LoggingApp
	Tracer    trace.Tracer
	Requester TopicRequester
	Nats      *nats.Conn
	Topic     string
}

func NewNatsPublishServer(app *LoggingApp, nc *nats.Conn, requester TopicRequester, topic string) *NatsPublishServer {
	tracer := otel.Tracer("NatsPublishServer")

	return &NatsPublishServer{
		Tracer:    tracer,
		App:       app,
		Nats:      nc,
		Requester: requester,
		Topic:     topic,
	}
}

func (s *NatsPublishServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	spanCtx, span := s.Tracer.Start(req.Context(), "ServeHTTP")
	defer span.End()
	requestData, err := dtos.ConvertRequestToQueueRequest(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.App.Logger.Error(err, "error converting request to queue data")
		http.Error(w, "can't convert request to queue data", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.String("requestId", requestData.RequestId))
	span.AddEvent("Marshal Event", trace.WithTimestamp(time.Now()))
	data, err := json.Marshal(requestData)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.App.Logger.Error(err, "error converting request to queue data", "requestId", requestData.RequestId)
		http.Error(w, "can't convert request to queue data", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.Int("queue_request_size", len(data)))
	span.AddEvent("request from queue")
	requestCtx, cancel := context.WithTimeout(spanCtx, time.Minute)
	defer cancel()
	responseHeaders, responseData, err := s.Requester.Request(requestCtx, span, data, s.Topic)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.App.Logger.Error(err, "failed to get queue response", "requestId", requestData.RequestId)
		http.Error(w, "failed to get queue response", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.Int("queue_response_size", len(responseData)))

	span.AddEvent("got response from queue")
	for headerKey, headerValue := range *responseHeaders {
		for _, headerSubValue := range headerValue {
			w.Header().Add(headerKey, headerSubValue)
		}
	}
	n, err := w.Write(responseData)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.App.Logger.Error(err, "error write response", err, "request_id", requestData.RequestId)
		http.Error(w, "can't write response", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.Int("response_size", n))
	span.SetStatus(codes.Ok, "")
}
