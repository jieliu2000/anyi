package test

import (
	"net/http"
	"net/http/httptest"
)

type MockServer struct {
	*httptest.Server

	RequestHandler func(w http.ResponseWriter, r *http.Request)
}

func (m *MockServer) Init() {
	m.Server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.RequestHandler != nil {
			m.RequestHandler(w, r)
		}
	}))
}

func (m *MockServer) SetRequestHandler(handler func(w http.ResponseWriter, r *http.Request)) {
	m.RequestHandler = handler
}

func (m *MockServer) URL() string {
	return m.Server.URL
}
func (m *MockServer) Start() {
	m.Server.Start()
}

func (m *MockServer) Close() {
	m.Server.Close()
}

func NewTestServer() *MockServer {

	mockServer := MockServer{}
	mockServer.Init()

	return &mockServer
}
