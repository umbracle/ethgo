package jsonrpc

import (
	"fmt"

	"github.com/umbracle/go-web3/jsonrpc/transport"
)

// Subscribe starts a new subscription
func (c *Client) Subscribe(method string, callback func(b []byte)) (func() error, error) {
	pub, ok := c.transport.(transport.PubSubTransport)
	if !ok {
		return nil, fmt.Errorf("Transport does not support the subscribe method")
	}
	close, err := pub.Subscribe(method, callback)
	return close, err
}
