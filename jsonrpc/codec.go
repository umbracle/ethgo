package jsonrpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Request is a jsonrpc request
type Request struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// Response is a jsonrpc response
type Response struct {
	ID     int             `json:"id"`
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

func encodeUintToHex(i uint64) string {
	return fmt.Sprintf("0x%x", i)
}

func parseUint64orHex(str string) (uint64, error) {
	base := 10
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
		base = 16
	}
	return strconv.ParseUint(str, base, 64)
}

func encodeToHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

func parseHexBytes(str string) ([]byte, error) {
	if !strings.HasPrefix(str, "0x") {
		return nil, fmt.Errorf("it does not have 0x prefix")
	}
	buf, err := hex.DecodeString(str[2:])
	if err != nil {
		return nil, err
	}
	return buf, nil
}

type BlockNumber int

const (
	Latest   BlockNumber = -1
	Earliest             = -2
	Pending              = -3
)

func (b BlockNumber) String() string {
	switch b {
	case Latest:
		return "latest"
	case Earliest:
		return "earliest"
	case Pending:
		return "pending"
	}
	if b < 0 {
		panic("internal. blocknumber is negative")
	}
	return fmt.Sprintf("0x%x", uint64(b))
}
