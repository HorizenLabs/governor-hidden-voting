package crypto

import (
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

// KeyPair is an election authority key-pair for ElGamal encryption of votes.
type KeyPair struct {
	Pk arith.CurvePoint // public key
	Sk arith.Scalar     // secret key
}

// NewKeyPair allocates and initializes a new valid ElGamal KeyPair.
func NewKeyPair(r io.Reader) (*KeyPair, error) {
	sk, pk, err := arith.RandomCurvePoint(r)
	if err != nil {
		return nil, err
	}
	keyPair := new(KeyPair)
	keyPair.Pk.Set(pk)
	keyPair.Sk.Set(sk)
	return keyPair, nil
}
