package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

//go:generate moq -out mock_connection.go . Connection
type Connection interface {
	ListenAndServe() error
	Addr() string
	ListenPath(path string, method string, handler http.HandlerFunc) error
	Shutdown(ctx context.Context)
}

type HTTPConnection struct {
	server *http.Server
	router *mux.Router
}

func NewHTTPConnection(bindAddr string, port int) Connection {
	r := mux.NewRouter()
	r.HandleFunc("/_health", healthHandler)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", bindAddr, port),
		Handler: r,
	}

	return &HTTPConnection{
		server: s,
		router: r,
	}
}

func (h *HTTPConnection) ListenAndServe() error {
	return h.server.ListenAndServe()
}

func (h *HTTPConnection) Addr() string {
	return h.server.Addr
}

func (h *HTTPConnection) ListenPath(path string, method string, handler http.HandlerFunc) error {
	h.router.HandleFunc(path, handler).Methods(method)

	return nil
}

func (h *HTTPConnection) Shutdown(ctx context.Context) {
	h.server.Shutdown(ctx)
}

func healthHandler(rw http.ResponseWriter, r *http.Request) {
	ioutil.ReadAll(r.Body)
	defer r.Body.Close()
}
