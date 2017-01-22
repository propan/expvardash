package main

import (
	"fmt"
	"net/url"
	"strings"
)

type Service struct {
	Name string
	URL  url.URL
}

func ParseURL(rawurl string) (*url.URL, error) {
	if !strings.HasPrefix(rawurl, "http") {
		rawurl = fmt.Sprintf("http://%s", rawurl)
	}

	url, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if url.Path == "" {
		url.Path = "/debug/vars"
	}
	return url, nil
}
