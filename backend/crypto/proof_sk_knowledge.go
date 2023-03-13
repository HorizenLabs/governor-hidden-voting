package crypto

import (
	"errors"
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

const numBytesProofSkKnowledge = arith.NumBytesCurvePoint +
	arith.NumBytesScalar

type ProofSkKnowledge struct {
	b arith.CurvePoint
	d arith.Scalar
}

func ProveSkKnowledge(reader io.Reader, keyPair *KeyPair) (*ProofSkKnowledge, error) {
	r, b, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}
	bytesPk, err := keyPair.Pk().MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesB, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	c := arith.FiatShamirChallenge(bytesPk, bytesB)
	d := new(arith.Scalar).Mul(c.Scalar(), keyPair.Sk())
	d = new(arith.Scalar).Add(r, d)
	proof := new(ProofSkKnowledge)
	proof.b.Set(b)
	proof.d.Set(d)
	return proof, nil
}

func VerifySkKnowledge(proof *ProofSkKnowledge, pk *arith.CurvePoint) error {
	m, err := proof.MarshalBinary()
	if err != nil {
		return err
	}
	proof = new(ProofSkKnowledge)
	err = proof.UnmarshalBinary(m)
	if err != nil {
		return err
	}

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB, err := proof.b.MarshalBinary()
	if err != nil {
		return err
	}
	c := arith.FiatShamirChallenge(bytesPk, bytesB)
	cPk := new(arith.CurvePoint).ScalarMult(pk, c.Scalar())
	bPlusCPk := new(arith.CurvePoint).Add(&proof.b, cPk)
	phi := new(arith.CurvePoint).ScalarBaseMult(&proof.d)
	if !phi.Equal(bPlusCPk) {
		return errors.New("sk knowledge proof verification failed")
	}
	return nil
}

func (proof *ProofSkKnowledge) MarshalBinary() ([]byte, error) {
	bytesB, err := proof.b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesD, err := proof.d.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ret := make([]byte, 0, numBytesProofSkKnowledge)
	ret = append(ret, bytesB...)
	ret = append(ret, bytesD...)
	return ret, nil
}

func (proof *ProofSkKnowledge) UnmarshalBinary(m []byte) error {
	if len(m) < numBytesProofSkKnowledge {
		return errors.New("ProofSkKnowledge: not enough data")
	}
	var err error

	if err = proof.b.UnmarshalBinary(m); err != nil {
		return err
	}
	m = m[arith.NumBytesCurvePoint:]
	if err = proof.d.UnmarshalBinary(m); err != nil {
		return err
	}

	return nil
}
