package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/antonholmquist/jason"
)

type Expvars struct {
	*jason.Object
}

type Fetcher interface {
	Fetch(url url.URL) (*Expvars, error)
}

type fetcher struct {
	client *http.Client
}

func NewFetcher() Fetcher {
	return &fetcher{
		client: &http.Client{
			Timeout: time.Second,
		},
	}
}

func (f *fetcher) Fetch(url url.URL) (*Expvars, error) {
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Could not fetch expvars from %s", url.String())
	}

	object, err := jason.NewObjectFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Expvars{object}, nil
}
