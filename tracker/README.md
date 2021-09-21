
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

var depositEvent = abi.MustNewEvent(`DepositEvent(
	bytes pubkey,
	bytes whitdrawalcred,
	bytes amount,
	bytes signature,
	bytes index
)`)

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

	store, err := boltdbStore.New("deposit.db")
	if err != nil {
		fmt.Printf("[ERR]: failted to start store %v", err)
		os.Exit(1)
	}

	tt, err := tracker.NewTracker(provider.Eth(),
		tracker.WithBatchSize(20000),
		tracker.WithStore(store),
		tracker.WithEtherscan(os.Getenv("ETHERSCAN_APIKEY")),
		tracker.WithFilter(&tracker.FilterConfig{
			Async: true,
			Address: []web3.Address{
				web3.HexToAddress(target),
			},
		}),
	)
	if err != nil {
		fmt.Printf("[ERR]: failed to create the tracker %v", err)
		os.Exit(1)
	}

	lastBlock, err := tt.GetLastBlock()
	if err != nil {
		fmt.Printf("[ERR]: failed to get last block %v", err)
		os.Exit(1)
	}
	if lastBlock != nil {
		fmt.Printf("Last block processed: %d\n", lastBlock.Number)
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		go func() {
			if err := tt.Sync(ctx); err != nil {
				fmt.Printf("[ERR]: %v", err)
			}
		}()

		go func() {
			for {
				select {
				case evnt := <-tt.EventCh:
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
				case <-tt.DoneCh:
					fmt.Println("historical sync done")
				}
			}
		}()

	}()

	handleSignals(cancelFn)
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

You can query the ETH2.0 Deposit contract like so:

```
go run main.go --endpoint https://mainnet.infura.io/v3/... --target 0x00000000219ab540356cbb839cbe05303d7705fa
```
