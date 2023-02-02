package http

import "net/http"

type ProxiesHttp interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}
