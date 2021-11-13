package jsonrpc

type Endpoints struct {
	Web3Client *Web3
	EthClient  *Eth
	Net        *Net
	Debug      *Debug
}

type IEndpoints interface {
	Eth() *Eth
	Web3() *Web3
}

func (r *Endpoints) Eth() *Eth {
	return r.EthClient
}

func (r *Endpoints) Web3() *Web3 {
	return r.Web3Client
}
