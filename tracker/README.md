
# Tracker

```
package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/jsonrpc"
	"github.com/umbracle/go-web3/tracker"

	boltdbStore "github.com/umbracle/go-web3/tracker/store/boltdb"
)

func main() {
	var endpoint string
	var target string

	flag.StringVar(&endpoint, "endpoint", "", "")
	flag.StringVar(&target, "target", "", "")

	flag.Parse()

	provider, err := jsonrpc.NewClient(endpoint)
	if err != nil {
		fmt.Printf("[ERR]: %v", err)
		os.Exit(1)
	}

	tConfig := tracker.DefaultConfig()
	tConfig.BatchSize = 2000
	tConfig.EtherscanFastTrack = true

	store, err := boltdbStore.New("db.db")
	if err != nil {
		fmt.Printf("[ERR]: %v", err)
		os.Exit(1)
	}

	t := tracker.NewTracker(provider.Eth(), tConfig)
	t.SetStore(store)

	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		if err := start(ctx, t, web3.HexToAddress(target)); err != nil {
			fmt.Printf("[ERR]: %v", err)
		}
	}()

	handleSignals(cancelFn)
}

var depositContract = `
tuple (
	pubkey bytes,
	whitdrawalcred bytes,
	amount bytes,
	signature bytes,
	index bytes
)
`

func start(ctx context.Context, t *tracker.Tracker, targetAddr web3.Address) error {
	if err := t.Start(ctx); err != nil {
		return err
	}

	fmt.Println("Tracker is ready")

	depositEvent := abi.NewEvent("DepositEvent", abi.MustNewType(depositContract))

	// create the filter
	fConfig := &tracker.FilterConfig{
		Async: true,
		Address: []web3.Address{
			targetAddr,
		},
	}
	f, err := t.NewFilter(fConfig)
	if err != nil {
		return err
	}

	go func() {
		for {
			evnt := <-f.EventCh
			for _, log := range evnt.Added {
				if depositEvent.Match(log) {
					vals, err := depositEvent.ParseLog(log)
					if err != nil {
						panic(err)
					}

					index := binary.LittleEndian.Uint64(vals["index"].([]byte))
					amount := binary.LittleEndian.Uint64(vals["amount"].([]byte))

					fmt.Printf("Deposit: Block %d Index %d Amount %d\n", log.BlockNumber, index, amount)
				}
			}
		}
	}()

	if err := f.Sync(ctx); err != nil {
		return err
	}

	fmt.Println("Historical sync is done")
	return nil
}

func handleSignals(cancelFn context.CancelFunc) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	<-signalCh

	gracefulCh := make(chan struct{})
	go func() {
		cancelFn()
		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return 1
	case <-gracefulCh:
		return 0
	}
}
```

You can query the Prysmatic labs Testnet ETH2.0 Deposit contract like so:

```
go run main.go --endpoint https://goerli.infura.io/v3/... --target 0x4689a3C63CE249355C8a573B5974db21D2d1b8Ef
```
