package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nicholasjackson/pipe/logger"
)

//go:generate moq -out mock_connection.go . Connection
type Connection interface {
	ListenAndServe() error
	Addr() string
	ListenPath(path string, method string, handler http.HandlerFunc) error
	CheckHealth() error
	Shutdown(ctx context.Context)
}

// HTTPConnection defines an HTTP Connection and is a wrapper for http.Server
type HTTPConnection struct {
	server *http.Server
	router *mux.Router
	log    logger.Logger
}

// NewHTTPConnection creates a new connection and bind to the given address and port
func NewHTTPConnection(bindAddr string, port int, logger logger.Logger) Connection {
	c := &HTTPConnection{}
	c.log = logger

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(c.notFoundHandler)
	r.HandleFunc("/_health", c.healthHandler)

	c.router = r

	c.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", bindAddr, port),
		Handler: r,
	}

	return c
}

// ListenAndServe calls the http.Server ListenAndServe method
func (h *HTTPConnection) ListenAndServe() error {
	return h.server.ListenAndServe()
}

// Addr returns the bound address for the connection
func (h *HTTPConnection) Addr() string {
	return h.server.Addr
}

// CheckHealth checks the health of the connection and returns an error if unhealthy
func (h *HTTPConnection) CheckHealth() error {
	resp, err := http.Get(fmt.Sprintf("http://%s/_health", h.Addr()))
	if err == nil && resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp != nil && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected status %d from health check, got status %d", http.StatusOK, resp.StatusCode)
	}

	return err
}

// ListenPath registers a new http.HandlerFunc at the given path
func (h *HTTPConnection) ListenPath(path string, method string, handler http.HandlerFunc) error {
	h.router.HandleFunc(path, handler).Methods(method)

	return nil
}

// Shutdown shuts down the connection
func (h *HTTPConnection) Shutdown(ctx context.Context) {
	h.server.Shutdown(ctx)
}

func (h *HTTPConnection) healthHandler(rw http.ResponseWriter, r *http.Request) {
	ioutil.ReadAll(r.Body)
	defer r.Body.Close()
}

func (h *HTTPConnection) notFoundHandler(rw http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	h.log.GetLogger().Error("Path not found", "url", r.URL.String(), "path", r.URL.RawPath, "data", string(data))
}
