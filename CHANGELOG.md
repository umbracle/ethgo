
# 0.1.2 (Unreleased)

- Add `testcases` for contract signature and transaction signing [[GH-193](https://github.com/umbracle/ethgo/issues/193)]
- Add `eth_feeHistory` rpc endpoint [[GH-192](https://github.com/umbracle/ethgo/issues/192)]
- Update `testserver` to `go-ethereum:v1.10.15` [[GH-191](https://github.com/umbracle/ethgo/issues/191)]
- Do not decode `to` in `Transaction` if not exists [[GH-190](https://github.com/umbracle/ethgo/issues/190)]

# 0.1.1 (25 April, 2022)

- Retrieve latest nonce when sending a transaction on `contract` [[GH-185](https://github.com/umbracle/ethgo/issues/185)]
- Add `etherscan.GasPrice` function to return last block gas price [[GH-182](https://github.com/umbracle/ethgo/issues/182)]
- Add `4byte` package and cli [[GH-178](https://github.com/umbracle/ethgo/issues/178)]
- Install and use `ethers.js` spec tests for wallet private key decoding [[GH-177](https://github.com/umbracle/ethgo/issues/177)]
- Add `GetLogs` function Etherscan to return logs by filter [[GH-170](https://github.com/umbracle/ethgo/issues/170)]
- Add `Copy` function to major data types [[GH-169](https://github.com/umbracle/ethgo/issues/169)]
- Parse `fixed bytes` type in event topic [[GH-168](https://github.com/umbracle/ethgo/issues/168)]
- Introduce `NodeProvider` and update `Contract` and `abigen` format. [[GH-167](https://github.com/umbracle/ethgo/issues/167)]

# 0.1.0 (5 March, 2022)

- Initial public release.
