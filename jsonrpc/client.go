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
	Call(method string, out interface{}, params ...interface{}) error
	Close() error
}

// NewClient creates a new client
func NewClient(addr string) (*Client, error) {
	c := &Client{}
	c.Endpoints = new(Endpoints)
	c.Endpoints.Web3Client = &Web3{c}
	c.Endpoints.EthClient = &Eth{c}
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

// EthClient returns the reference to the eth namespace
func (c *Client) Eth() *Eth {
	return c.Endpoints.EthClient
}

// Net returns the reference to the net namespace
func (c *Client) Net() *Net {
	return c.Endpoints.Net
}

// Web3Client returns the reference to the web3 namespace
func (c *Client) Web3() *Web3 {
	return c.Endpoints.Web3Client
}

// EthClient returns the reference to the eth namespace
func (c *Client) Debug() *Debug {
	return c.Endpoints.Debug
}
