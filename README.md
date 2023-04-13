This repository implements a proof of concept for private voting on blockchain.
The protocol is based on the Helios voting system ([https://eprint.iacr.org/2016/765.pdf](https://eprint.iacr.org/2016/765.pdf)) with a few modifications:
- an EVM compatible chain is used both as a public bulletin board to cast encrypted votes and as a zk-proof checker
- EC ElGamal encryption is used. The bn256 curve is used, chiefly because the EVM has precompiled contracts for efficiently executing arithmetic on this curve

The proof of concept consists of:
- [backend](./backend/), a cryptographic backend written in Go, which implements functions to perform encryption/decryption and zk-proof generation
- [smart-contracts](./smart-contracts/), containing a Solidity smart contract which implements the voting logic (including on-chain zk-proof checking) 

For the moment, the proof of concept has the following limitations:
- only yes-no voting is supported
- a single authority is responsible for performing tallying
- there is no voter authentication (i.e. anyone can vote an arbitrary number of times by creating new Ethereum accounts)
