# Private on-chain voting
This repository implements a proof of concept for private voting on blockchain.
The protocol is based on the Helios voting system ([https://eprint.iacr.org/2016/765.pdf](https://eprint.iacr.org/2016/765.pdf)) with a few modifications:
- an EVM compatible chain is used both as a public bulletin board to cast encrypted votes and as a zk-proof checker
- EC ElGamal encryption is adopted. The bn256 curve is used, chiefly because the EVM has precompiled contracts for efficiently executing arithmetic on this curve

## Repo structure
- [`backend`](./backend/), a cryptographic backend written in Go, which implements functions for:
    * generating keypairs for encryption/decryption
    * performing encryption/decryption of votes
    * performing homomorphic addition and scaling of encrypted votes
    * generating all the zk-proofs required by the protocol

  The functionality of the Go backend is accessible:
    * directly via the Go modules [`crypto`](./backend/crypto/) and [`arith`](./backend/arith/)
    * as a WebAssembly instance in [`wasm`](./backend/wasm/)
- [`smart-contracts/contracts`](./smart-contracts/), a set of Solidity smart contracts
    * [`cryptography`](./smart-contracts/contracts/cryptography/) contains a contract to verify the zk-proofs required by the protocol.
    * [`openzeppelin-voting`](./smart-contracts/contracts/openzeppelin-voting/) contains a set of contracts which allow to deploy private voting as an extension of [OpenZeppelin governance framework](https://docs.openzeppelin.com/contracts/4.x/api/governance).
    * [`simple-voting`](./smart-contracts/contracts/simple-voting/) contains a basic smart contract to show in a simple way how the Helios protocol works in an on-chain setting.

## Cloning
The repo contains [openzeppelin-contracts](https://github.com/OpenZeppelin/openzeppelin-contracts) repo as a submodule (for ease of testing).
Cloning should be done via the command
```
git clone --recurse-submodules https://github.com/HorizenLabs/e-voting-poc.git
```
If cloning has already been performed without the `--recurse-submodules` option, the directory `smart-contracts/lib/openzeppelin-contracts` should be there, but empty. In order to populate it, please issue the command
```
git submodule update --init --recursive
```

## Dependencies
- The cryptographic backend needs a working [Go installation](https://go.dev/doc/install)
- Running javascript smart contract tests requires [Node.js and npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)

## Building
The javascript tests of the smart contracts invoke the Go backend via the wasm wrapper, which can be built by going to the `backend/wasm` directory and issuing the command
```
make
```
Moreover, hardhat should be installed by going to the `smart-contracts` directory and issuing the command
```
npm install --save-dev hardhat
```

## Testing
To run all the Go backend tests, switch to `backend` directory and issue the command
```
go test `go list ./... | grep -v wasm/cmd/wasm` -v
```
This excludes the files inside `wasm/cmd/wasm` directory, which should be compiled in wasm. To run the tests of that module (which currently has none) it would be necessary to issue the command
```
GOOS=js GOARCH=wasm go test ./wasm/cmd/wasm/ -v
```

To run all the Solidity smart contract tests, switch to `smart-contracts` directory and issue the command
```
npx hardhat test
```
Before doing so, it's necessary to build the wasm backend as described in the [Building](#building) section, and populating the `openzeppelin-contracts` submodule as described in the [Cloning](#cloning) section.

## Limitations
For the moment, the implementation has the following limitations:
- only yes-no voting is supported
- a single authority is responsible for performing tallying

## Security
This code is experimental, and has not yet been audited. Using it in a production system is not recommended.
