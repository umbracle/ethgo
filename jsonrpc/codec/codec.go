package codec

import (
	"encoding/json"
	"fmt"
)

// Request is a jsonrpc request
type Request struct {
	ID     uint64          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// Response is a jsonrpc response
type Response struct {
	ID     uint64          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *ErrorObject    `json:"error,omitempty"`
}

// ErrorObject is a jsonrpc error
type ErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements error interface
func (e *ErrorObject) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("jsonrpc.internal marshal error: %v", err)
	}
	return string(data)
}
