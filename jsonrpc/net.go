package jsonrpc

// Net is the net namespace
type Net struct {
	c *Client
}

// Version returns the current network id
func (n *Net) Version() (uint64, error) {
	var out string
	if err := n.c.Call("net_version", &out); err != nil {
		return 0, err
	}
	return parseUint64orHex(out)
}

// Listening returns true if client is actively listening for network connections
func (n *Net) Listening() (bool, error) {
	var out bool
	err := n.c.Call("net_listening", &out)
	return out, err
}

// PeerCount returns number of peers currently connected to the client
func (n *Net) PeerCount() (uint64, error) {
	var out string
	if err := n.c.Call("net_peerCount", &out); err != nil {
		return 0, err
	}
	return parseUint64orHex(out)
}
