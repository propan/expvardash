package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/stretchr/testify/assert"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

func tearUp() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
}

func tearDown() {
	server.Close()
}

func ParseTestURL(t *testing.T, rawurl string) *url.URL {
	url, err := url.Parse(rawurl)
	assert.Nil(t, err)
	return url
}

func TestFetcher_Fetch_Timeout(t *testing.T) {
	tearUp()
	defer tearDown()

	mux.HandleFunc("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		fmt.Fprint(w, "OK")
	})

	f := NewFetcher()
	f.(*fetcher).client.Timeout = 10 * time.Microsecond

	_, err := f.Fetch(*ParseTestURL(t, server.URL+"/debug/vars"))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "(Client.Timeout exceeded while awaiting headers)")
}

func TestFetcher_Fetch_BadStatusCode(t *testing.T) {
	tearUp()
	defer tearDown()

	mux.HandleFunc("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, "ERROR")
	})

	f := NewFetcher()

	_, err := f.Fetch(*ParseTestURL(t, server.URL+"/debug/vars"))
	assert.Equal(t, errors.New("Could not fetch expvars from "+server.URL+"/debug/vars"), err)
}

func TestFetcher_Fetch_BadResponse(t *testing.T) {
	tearUp()
	defer tearDown()

	mux.HandleFunc("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{")
	})

	f := NewFetcher()

	_, err := f.Fetch(*ParseTestURL(t, server.URL+"/debug/vars"))
	assert.Equal(t, errors.New("unexpected EOF"), err)
}

func TestFetcher_Fetch_Success(t *testing.T) {
	tearUp()
	defer tearDown()

	expvarsResponse := `{"gauge": {"metric": 800}, "process": {"text": "text 1"}, "memstats": {"alloc": 123}}`
	o, err := jason.NewObjectFromBytes([]byte(expvarsResponse))
	assert.NoError(t, err)

	mux.HandleFunc("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expvarsResponse)
	})

	f := NewFetcher()

	vars, err := f.Fetch(*ParseTestURL(t, server.URL+"/debug/vars"))
	assert.Nil(t, err)
	assert.Equal(t, &Expvars{o}, vars)
}
