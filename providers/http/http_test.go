package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/matryer/is"
)

var serverResponse []byte

func setupHTTPProvider(t *testing.T) (*is.I, *HTTPProvider, func()) {
	is := is.New(t)

	httptest := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			d, err := ioutil.ReadAll(r.Body)
			is.NoErr(err) // expected no error from http body

			serverResponse = d
			rw.Write([]byte("ok"))
		}),
	)

	u, _ := url.Parse(httptest.URL)
	port, _ := strconv.Atoi(u.Port())

	p := &HTTPProvider{
		Protocol: u.Scheme,
		Server:   u.Hostname(),
		Port:     port,
		Path:     u.Path,
	}

	stats, _ := statsd.New("http://localhost:8125")
	p.Setup(nil, hclog.Default(), stats)

	return is, p, func() {
		httptest.Close()
	}
}

func TestPublishCallsEndpointWithData(t *testing.T) {
	is, p, cleanup := setupHTTPProvider(t)
	defer cleanup()
	payload := []byte("test data")

	_, err := p.Publish(payload)

	is.NoErr(err)                     // should not have reuturned an error
	is.Equal(payload, serverResponse) // should have sent the correct payload
}

func TestPublishCallsEndpointAndReturnsBody(t *testing.T) {
	is, p, cleanup := setupHTTPProvider(t)
	defer cleanup()
	payload := []byte("test data")

	data, err := p.Publish(payload)

	is.NoErr(err)                // should not have reuturned an error
	is.Equal("ok", string(data)) // should have sent the correct payload
}
