package transport

import "strings"

// Transport is an inteface for transport methods to send jsonrpc requests
type Transport interface {
	// Call makes a jsonrpc request
	Call(method string, out interface{}, params ...interface{}) error

	// Close closes the transport connection if necessary
	Close() error
}

// NewTransport creates a new transport object
func NewTransport(url string) (Transport, error) {
	if strings.HasPrefix(url, "ws://") {
		return newWebsocket(url)
	}
	return newHTTP(url), nil
}
