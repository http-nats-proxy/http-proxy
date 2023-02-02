package dtos

import (
	"bytes"
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestConvertRequestToQueueRequestData(t *testing.T) {
	url := "some_url"
	bodyData := []byte("some_body")
	bodyBuffer := bytes.NewBuffer(bodyData)
	httpRequest, _ := http.NewRequest(http.MethodGet, url, bodyBuffer)
	response, err := ConvertRequestToQueueRequestData(httpRequest)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(response.Body, bodyData) != 0 {
		t.Errorf("expected %s, got %s", string(bodyData), string(response.Body))
	}
	if response.Method != httpRequest.Method {
		t.Errorf("expected %s, got %s", http.MethodGet, response.Method)
	}
	if !reflect.DeepEqual(response.Header, response.Headers) {
		t.Errorf("Header: expected %s, got %s", httpRequest.Header, response.Header)
	}

}

func TestConvertRequestToQueueRequestWithoutRequestId(t *testing.T) {
	url := "some_url"
	bodyData := []byte("some_body")
	bodyBuffer := bytes.NewBuffer(bodyData)
	httpRequest, _ := http.NewRequest(http.MethodGet, url, bodyBuffer)
	response, err := ConvertRequestToQueueRequest(httpRequest)
	if err != nil {
		t.Error(err)
	}
	if response.RequestId == "" {
		t.Errorf("expected request id, got none")
	}
}
func TestConvertRequestToQueueRequestWithRequestId(t *testing.T) {
	url := "some_url"
	bodyData := []byte("some_body")
	bodyBuffer := bytes.NewBuffer(bodyData)
	httpRequest, _ := http.NewRequest(http.MethodGet, url, bodyBuffer)
	httpRequest.Header.Set("x-request-id", "some_request_id")
	response, err := ConvertRequestToQueueRequest(httpRequest)
	if err != nil {
		t.Error(err)
	}
	if response.RequestId == "" {
		t.Errorf("expected request id, got none")
	}
}
func TestConvertRequestToQueueRequestPassesContext(t *testing.T) {
	url := "some_url"
	bodyData := []byte("some_body")
	bodyBuffer := bytes.NewBuffer(bodyData)
	ctx := context.WithValue(context.Background(), "some_key", "some_value")
	httpRequest, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, bodyBuffer)
	httpRequest.Header.Set("x-request-id", "some_request_id")
	response, err := ConvertRequestToQueueRequest(httpRequest)
	if err != nil {
		t.Error(err)
	}
	if response.Ctx.Value("some_key") != "some_value" {
		t.Errorf("expected %s, got %s", "some_value", response.Ctx.Value("some_key"))
	}
}
