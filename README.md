# Governor hidden voting
Horizen Labs developed an on-chain private voting solution which can be deployed on EVM-compatible blockchains, and is compatible with the widely adopted [OpenZeppelin Governor framework](https://docs.openzeppelin.com/contracts/4.x/api/governance).

An overview of the protocol:
- Votes are encrypted using a **linear homomorphic encryption scheme** (El-Gamal instantiated on curve bn254 to be precise).
- **Only Yes/No votes are considered**, even though it would be feasible to extend the protocol to more general ballot types.
- **Our protocol relies on a tallying authority**, an entity who possesses some private information which enables them to perform tallying.

The solution migrates [Helios voting protocol](https://eprint.iacr.org/2016/765.pdf) to on-chain, using linear homomorphic encryption of the votes and efficient zk-proofs to guarantee:
- integrity of the voting process
- privacy of the votes
- censorship resistance
- reasonable gas costs

For additional details, please visit our [blog post](https://hackmd.io/@hackmdhl/BJSz8pnan).

## Description of the protocol
The rough idea is the following:
- Users cast their vote in encrypted form, using a linearly-homomorphic, asymmetric encryption scheme (EC ElGamal). The encryption key is public, while the decryption key is only known to the tallying authority.
- To ensure that the vote is valid (i.e. it encodes a 0 (Against) or a 1 (For)) without requiring decryption, users should also send a zk-proof of vote well-formedness together with their vote.
- The encrypted votes and zk-proofs of vote well-formedness are received by a smart contract. If the zk-proof is invalid, then the vote is discarded. Otherwise, the smart contract scales the encrypted vote by the voting power of the respective voter, and then accumulates the result into a running total. This is possible thanks to the homomorphic property of the encryption scheme.
- At the end of the voting period, the tallying authority, using its secret decryption key, decrypts the accumulated result and posts it on-chain, together with a zk-proof of correct decryption.

The tallying authority must be trusted because, even if the protocol prevents it from forging invalid votes, or censoring legitimate votes, it can still:
- Decrypt any individual vote. This could be mitigated by implementing a threshold decryption scheme and appointing multiple independent tallying authorities, so that no individual entity can decrypt single votes. The cryptographic backend does not yet implement this feature, but it is described in [https://eprint.iacr.org/2016/765.pdf](https://eprint.iacr.org/2016/765.pdf)
- Refuse to perform tallying, making it impossible to know the final result. This could be mitigated by putting in place some crypto-economic deterrent (e.g. requiring to put up some collateral to become a tallying authority, and slashing it in case of malicious behavior).
## Compatibility with OpenZeppelin Governor
The contracts contained inside [smart-contracts/contracts/openzeppelin-voting](smart-contracts/contracts/openzeppelin-voting) allow to deploy the proposed private voting solution as a Governor contract, using the modular OpenZeppelin governance framework.

A DAO wanting to implement private voting would have to:
- Update its Governor contract to use our GovernorEncrypted module (see the [GovernorEncryptedMock](smart-contracts/contracts/mocks/GovernorEncryptedMock.sol) contract for a concrete example).
- Appoint a trusted entity as tallying authority, whose duty is to decrypt the result at the end of the voting period (see [openzeppelin-voting readme](smart-contracts/contracts/openzeppelin-voting/README.md) for an explanation of the lifecycle of a proposal).
- Develop a frontend for users to interact with the contract. Unfortunately [Tally](https://tally.xyz), the most used frontend for Governor-like governance contracts, is not compatible with our solution. This is unavoidable, since private voting is a novelty in the field of on-chain governance. The [wasm module](backend/wasm) should make it easy to interact with the cryptographic backend directly from a browser.

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
This code has not yet been audited, use it at your own risk.
