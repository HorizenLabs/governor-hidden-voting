package crypto

import (
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

type KeyPair struct {
	Pk arith.CurvePoint
	Sk arith.Scalar
}

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
