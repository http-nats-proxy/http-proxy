package dtos

import (
	"context"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
)

type QueueRequestData struct {
	Headers http.Header `json:"headers"`

	Method           string      `json:"method"`
	URL              string      `json:"url"`
	ProtoMajor       int         `json:"protoMajor"`
	ProtoMinor       int         `json:"protoMinor"`
	Header           http.Header `json:"header"`
	Body             []byte      `json:"body"`
	ContentLength    int64       `json:"contentLength"`
	TransferEncoding []string    `json:"transferEncoding"`
	Host             string      `json:"host"`
	RemoteAddr       string      `json:"remoteAddr"`
	RequestURI       string      `json:"requestURI"`
}

func ConvertRequestToQueueRequestData(req *http.Request) (*QueueRequestData, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return nil, err
	}
	//requestId := uuid.New().String()
	//if req.Header.Get("X-Request-Id") != "" {
	//	requestId = req.Header.Get("X-Request-Id")
	//}

	return &QueueRequestData{
		Headers:          req.Header,
		Method:           req.Method,
		URL:              req.URL.String(),
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Header:           req.Header,
		Body:             body,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Host:             req.Host,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
	}, nil

}

type QueueRequest struct {
	Kind      string
	Data      QueueRequestData
	ctx       context.Context
	RequestId string
}

func ConvertRequestToQueueRequest(req *http.Request) (*QueueRequest, error) {
	data, err := ConvertRequestToQueueRequestData(req)
	if err != nil {
		return nil, err
	}
	requestId := uuid.New().String()
	if req.Header.Get("X-Request-Id") != "" {
		requestId = req.Header.Get("X-Request-Id")
	}
	return &QueueRequest{
		Data:      *data,
		Kind:      "QueueRequestData",
		ctx:       req.Context(),
		RequestId: requestId,
	}, nil

}
