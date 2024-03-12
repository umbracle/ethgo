package gaspricer

import "context"

type GasPricer interface {
	Start(ctx context.Context)
	GasPrice() uint64
}
