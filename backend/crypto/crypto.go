// Package crypto implements the cryptographic algorithms used in the
// Helios e-voting protocol.
//
// The implementation closely follows [Cryptographic Voting], except
// that ElGamal cryptosystem is instantiated using elliptic curves
// instead of integer factorization.
//
// At the moment the package has the following limitations:
//   - only yes-no votes are supported (no multi-choice voting)
//   - tallying is performed by a single entity (no threshold decryption)
//
// [Cryptographic Voting]: https://eprint.iacr.org/2016/765.pdf
package crypto
