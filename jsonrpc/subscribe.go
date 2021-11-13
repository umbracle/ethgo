package jsonrpc

import (
	"fmt"

	"github.com/panyanyany/go-web3/jsonrpc/transport"
)

// SubscriptionEnabled returns true if the subscription Endpoints are enabled
func (c *Client) SubscriptionEnabled() bool {
	_, ok := c.Transport.(transport.PubSubTransport)
	return ok
}

// Subscribe starts a new subscription
func (c *Client) Subscribe(method string, callback func(b []byte)) (func() error, error) {
	pub, ok := c.Transport.(transport.PubSubTransport)
	if !ok {
		return nil, fmt.Errorf("Transport does not support the subscribe method")
	}
	close, err := pub.Subscribe(method, callback)
	return close, err
}
