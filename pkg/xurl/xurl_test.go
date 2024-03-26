package xurl

import (
	"net/url"
	"testing"

	"github.com/ignite/cli/v28/ignite/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		addr string
		want *url.URL
		err  error
	}{
		{
			name: "http",
			addr: "http://localhost",
			want: &url.URL{Host: "localhost:80", Scheme: schemeHTTP},
		},
		{
			name: "https",
			addr: "https://localhost",
			want: &url.URL{Host: "localhost:443", Scheme: schemeHTTPS},
		},
		{
			name: "custom",
			addr: "http://localhost:4000",
			want: &url.URL{Host: "localhost:4000", Scheme: schemeHTTP},
		},
		{
			name: "custom ssl",
			addr: "https://localhost:4005",
			want: &url.URL{Host: "localhost:4005", Scheme: schemeHTTPS},
		},
		{
			name: "no schema and port",
			addr: "localhost",
			want: &url.URL{Host: "localhost:80", Scheme: schemeHTTP},
		},
		{
			name: "no schema",
			addr: "localhost:80",
			want: &url.URL{Host: "localhost:80"},
		},
		{
			name: "invalid address",
			addr: "://.e",
			err:  errors.New("parse \"://.e\": missing protocol scheme"),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := Parse(tt.addr)
			if tt.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.err.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tt.want, addr)
		})
	}
}

func Test_addressPort(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		wantHost string
		want     bool
	}{
		{
			name: "URI path",
			arg:  "/test/false",
			want: false,
		},
		{
			name: "invalid address",
			arg:  "aeihf3/aef/f..//",
			want: false,
		},
		{
			name:     "host and port",
			arg:      "102.33.3.43:10000",
			wantHost: "102.33.3.43:10000",
			want:     true,
		},
		{
			name:     "local port",
			arg:      "0.0.0.0:10000",
			wantHost: "0.0.0.0:10000",
			want:     true,
		},
		{
			name:     "only port",
			arg:      ":10000",
			wantHost: "0.0.0.0:10000",
			want:     true,
		},
		{
			name: "only host",
			arg:  "102.33.3.43",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, got := addressPort(tt.arg)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.wantHost, gotHost)
		})
	}
}

func Test_address(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{
			name:    "localhost",
			address: "localhost",
			want:    "localhost",
		},
		{
			name:    "empty port",
			address: "127.0.0.1",
			want:    "127.0.0.1",
		},
		{
			name:    "empty string",
			address: "",
			want:    "",
		},
		{
			name:    "empty host and port",
			address: ":",
			want:    "localhost:",
		},
		{
			name:    "empty host",
			address: ":80",
			want:    "localhost:80",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := address(tt.address)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsSSL(t *testing.T) {
	tests := []struct {
		name string
		url  *url.URL
		want bool
	}{
		{
			name: "no ssl",
			url:  &url.URL{Host: "localhost:80", Scheme: schemeHTTP},
			want: false,
		},
		{
			name: "just not ssl schema",
			url:  &url.URL{Scheme: schemeHTTP},
			want: false,
		},
		{
			name: "ssl",
			url:  &url.URL{Host: "localhost:80", Scheme: schemeHTTPS},
			want: true,
		},
		{
			name: "just ssl schema",
			url:  &url.URL{Scheme: schemeHTTPS},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSSL(tt.url)
			require.Equal(t, tt.want, got)
		})
	}
}
