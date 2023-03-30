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

import (
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

func NewKeyPairWithProof(r io.Reader) (*KeyPair, *ProofSkKnowledge, error) {
	keyPair, err := NewKeyPair(r)
	if err != nil {
		return nil, nil, err
	}
	proof, err := ProveSkKnowledge(r, keyPair)
	if err != nil {
		return nil, nil, err
	}
	return keyPair, proof, nil
}

func EncryptVoteWithProof(r io.Reader, vote int64, pk *arith.CurvePoint) (*EncryptedVote, *ProofVoteWellFormedness, error) {
	encryptedVote, secret, err := Vote(vote).Encrypt(r, pk)
	if err != nil {
		return nil, nil, err
	}
	proof, err := ProveVoteWellFormedness(r, encryptedVote, Vote(vote), secret, pk)
	if err != nil {
		return nil, nil, err
	}
	return encryptedVote, proof, nil
}

func DecryptTallyWithProof(r io.Reader, tally *EncryptedVote, n int64, keyPair *KeyPair) (int64, *ProofCorrectDecryption, error) {
	decryptedVote, err := tally.Decrypt(&keyPair.Sk, n)
	if err != nil {
		return -1, nil, err
	}
	proof, err := ProveCorrectDecryption(r, tally, keyPair)
	if err != nil {
		return -1, nil, err
	}
	return int64(decryptedVote), proof, nil
}
