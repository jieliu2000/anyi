package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func NewTestServer() *httptest.Server {

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.Proto)
	}))

	return ts
}
