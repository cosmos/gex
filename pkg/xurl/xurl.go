package xurl

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

// Parse ensures that url has a port number and scheme suits with the connection type.
func Parse(s string) (*url.URL, error) {
	// Check if the URL contains a schema
	if !strings.Contains(s, "://") {
		// Handle the case where the URI is an IP:PORT or HOST:PORT
		// without scheme prefix because that case can't be URL parsed.
		// When the URI has no scheme it is parsed as a path by "url.Parse"
		// placing the colon within the path, which is invalid.
		if host, isAddrPort := addressPort(address(s)); isAddrPort {
			return &url.URL{Host: host}, nil
		}

		// Prepend a default schema (e.g., "http://") if it doesn't have one
		s = fmt.Sprintf("%s://%s", schemeHTTP, s)
	}

	// Parsing the URL
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	port := u.Port()
	if port == "" {
		port = "80"
		if u.Scheme == schemeHTTPS {
			port = "443"
		}
	}

	u.Host = fmt.Sprintf("%s:%s", u.Hostname(), port)
	return u, nil
}

// IsSSL ensures that address is SSL protocol.
func IsSSL(url *url.URL) bool {
	return url.Scheme == schemeHTTPS
}

// address ensures that address contains localhost as host if non specified.
func address(address string) string {
	if strings.HasPrefix(address, ":") {
		return "localhost" + address
	}
	return address
}

// addressPort verify if the string is an address and port host.
func addressPort(s string) (string, bool) {
	// Use the net split function to support IPv6 addresses
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return "", false
	}
	if host == "" {
		host = "0.0.0.0"
	}
	return net.JoinHostPort(host, port), true
}
