package crypto

import (
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

type KeyPair struct {
	pk arith.CurvePoint
	sk arith.Scalar
}

func (keyPair *KeyPair) Pk() *arith.CurvePoint {
	return &keyPair.pk
}

func (keyPair *KeyPair) Sk() *arith.Scalar {
	return &keyPair.sk
}

func NewKeyPair(r io.Reader) (*KeyPair, error) {
	sk, pk, err := arith.RandomCurvePoint(r)
	if err != nil {
		return nil, err
	}
	return &KeyPair{pk: *pk, sk: *sk}, nil
}
