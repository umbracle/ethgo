package jsonrpc

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type Client struct {
	addr      string
	client    *fasthttp.Client
	endpoints endpoints
}

type endpoints struct {
	w *Web3
	e *Eth
	n *Net
}

func NewClient(addr string) *Client {
	c := &Client{
		addr:   addr,
		client: &fasthttp.Client{},
	}
	c.endpoints.w = &Web3{c}
	c.endpoints.e = &Eth{c}
	c.endpoints.n = &Net{c}
	return c
}

func (c *Client) Call(method string, out interface{}, params ...interface{}) error {
	// Encode json-rpc request
	request := Request{
		Method: method,
	}
	if len(params) > 0 {
		data, err := json.Marshal(params)
		if err != nil {
			panic(err)
		}
		request.Params = data
	}
	raw, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(c.addr)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody(raw)

	if err := c.client.Do(req, res); err != nil {
		return err
	}

	// Decode json-rpc response
	var response Response
	if err := json.Unmarshal(res.Body(), &response); err != nil {
		return err
	}
	if response.Error != nil {
		return response.Error
	}
	if err := json.Unmarshal(response.Result, out); err != nil {
		return err
	}
	return nil
}
