#!/bin/bash

# Start geth test server
docker run --name geth-test -d -p 8545:8545 -p 8546:8546 ethereum/client-go:v1.10.15 \
    --dev \
    --datadir /eth1data \
    --ipcpath /eth1data/geth.ipc \
    --http --http.addr 0.0.0.0  --http.vhosts=* --http.api eth,net,web3,debug \
    --ws --ws.addr 0.0.0.0 \
    --verbosity 4

# Wait for geth to be running
while ! nc -z localhost 8545; do   
  sleep 0.1 # wait for 1/10 of the second before check again
done
