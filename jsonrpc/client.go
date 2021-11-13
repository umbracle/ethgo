package jsonrpc

import (
	"github.com/panyanyany/go-web3/jsonrpc/transport"
)

// Client is the jsonrpc client
type Client struct {
	Transport transport.Transport
	Endpoints *Endpoints
}

type IClient interface {
	Eth() *Eth
	Call(method string, out interface{}, params ...interface{}) error
	Close() error
}

type Endpoints struct {
	Web3  *Web3
	Eth   *Eth
	Net   *Net
	Debug *Debug
}

// NewClient creates a new client
func NewClient(addr string) (*Client, error) {
	c := &Client{}
	c.Endpoints = new(Endpoints)
	c.Endpoints.Web3 = &Web3{c}
	c.Endpoints.Eth = &Eth{c}
	c.Endpoints.Net = &Net{c}
	c.Endpoints.Debug = &Debug{c}

	t, err := transport.NewTransport(addr)
	if err != nil {
		return nil, err
	}
	c.Transport = t
	return c, nil
}

// Close closes the tranport
func (c *Client) Close() error {
	return c.Transport.Close()
}

// Call makes a jsonrpc call
func (c *Client) Call(method string, out interface{}, params ...interface{}) error {
	return c.Transport.Call(method, out, params...)
}
