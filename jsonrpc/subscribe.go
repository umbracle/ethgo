package jsonrpc

import (
	"context"
	"fmt"

	"github.com/umbracle/ethgo/jsonrpc/transport"
)

// SubscriptionEnabled returns true if the subscription endpoints are enabled
func (c *Client) SubscriptionEnabled() bool {
	_, ok := c.transport.(transport.PubSubTransport)
	return ok
}

// Subscribe starts a new subscription
func (c *Client) Subscribe(ctx context.Context, method string, callback func(b []byte)) error {
	pub, ok := c.transport.(transport.PubSubTransport)
	if !ok {
		return fmt.Errorf("transport does not support the subscribe method")
	}
	return pub.Subscribe(ctx, method, callback)
}
