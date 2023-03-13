package crypto

import (
	"errors"
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

const numBytesKeyPair = arith.NumBytesCurvePoint + arith.NumBytesScalar

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

func (keyPair *KeyPair) MarshalBinary() ([]byte, error) {
	pk, err := keyPair.pk.MarshalBinary()
	if err != nil {
		return nil, err
	}
	sk, err := keyPair.sk.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ret := make([]byte, 0, numBytesKeyPair)
	ret = append(ret, pk...)
	ret = append(ret, sk...)
	return ret, nil
}

func (keyPair *KeyPair) UnmarshalBinary(m []byte) error {
	if len(m) < numBytesKeyPair {
		return errors.New("KeyPair: not enough data")
	}
	err := keyPair.pk.UnmarshalBinary(m)
	if err != nil {
		return err
	}
	m = m[arith.NumBytesCurvePoint:]
	err = keyPair.sk.UnmarshalBinary(m)
	if err != nil {
		return err
	}
	return nil
}
