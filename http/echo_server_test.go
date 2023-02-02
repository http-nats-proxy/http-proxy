package http

import (
	"bytes"
	"encoding/json"
	"github.com/http-nats-proxy/http-proxy/dtos"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEchoServer(t *testing.T) {
	url := "http://someurl/foo"
	bodyData := []byte("some_body")
	server := NewEchoServer()

	bodyBuffer := bytes.NewBuffer(bodyData)

	req := httptest.NewRequest("POST", url, bodyBuffer)
	req.Header.Set("Some-Key", "Some-Value")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)

	}

	var data dtos.QueueRequestData
	err := json.Unmarshal(body, &data)
	if err != nil {
		t.Error(err)
	}
	if data.Headers.Get("Some-Key") != "Some-Value" {
		t.Errorf("Expected value %s, got %s", "Some-Value", data.Headers.Get("Some-Key"))
	}
	if bytes.Compare(data.Body, bodyData) != 0 {
		t.Errorf("Expected body %s, got %s", string(bodyData), string(data.Body))
	}
	if data.Method != "POST" {
		t.Errorf("Expected method %s, got %s", "POST", data.Method)
	}
}
