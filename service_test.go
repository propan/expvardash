package main

import (
	"net/url"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		rawurl  string
		want    *url.URL
		wantErr error
	}{
		{
			name:    "bad url",
			rawurl:  "http://^",
			wantErr: errors.New(`parse http://^: invalid character "^" in host name`),
		},
		{
			name:   "url without transport",
			rawurl: "localhost:5678/test",
			want:   ParseTestURL(t, "http://localhost:5678/test"),
		},
		{
			name:   "url without path",
			rawurl: "localhost:5678",
			want:   ParseTestURL(t, "http://localhost:5678/debug/vars"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(tt.rawurl)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, got)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}
