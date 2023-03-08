package crypto

import (
	"errors"
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

const numBytesProofSkKnowledge = arith.NumBytesCurvePoint +
	arith.NumBytesScalar

type ProofSkKnowledge struct {
	b *arith.CurvePoint
	d *arith.Scalar
}

func ProveSkKnowledge(reader io.Reader, keyPair *KeyPair) (*ProofSkKnowledge, error) {
	r, b, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}
	c := arith.FiatShamirChallenge(
		keyPair.Pk().Marshal(),
		b.Marshal())
	d := new(arith.Scalar).Mul(c.Scalar(), keyPair.Sk())
	d = new(arith.Scalar).Add(r, d)
	return &ProofSkKnowledge{b: b, d: d}, nil
}

func VerifySkKnowledge(proof *ProofSkKnowledge, pk *arith.CurvePoint) error {
	m := proof.Marshal()
	proof = new(ProofSkKnowledge)
	proof.Unmarshal(m)

	c := arith.FiatShamirChallenge(
		pk.Marshal(),
		proof.b.Marshal())
	cPk := new(arith.CurvePoint).ScalarMult(pk, c.Scalar())
	bPlusCPk := new(arith.CurvePoint).Add(proof.b, cPk)
	phi := new(arith.CurvePoint).ScalarBaseMult(proof.d)
	if !phi.Equal(bPlusCPk) {
		return errors.New("sk knowledge proof verification failed")
	}
	return nil
}

func (proof *ProofSkKnowledge) Marshal() []byte {
	ret := make([]byte, 0, numBytesProofSkKnowledge)
	ret = append(ret, proof.b.Marshal()...)
	ret = append(ret, proof.d.Marshal()...)
	return ret
}

func (proof *ProofSkKnowledge) Unmarshal(m []byte) ([]byte, error) {
	if len(m) < numBytesProofSkKnowledge {
		return nil, errors.New("ProofSkKnowledge: not enough data")
	}
	var err error

	if proof.b == nil {
		proof.b = &arith.CurvePoint{}
	}
	if m, err = proof.b.Unmarshal(m); err != nil {
		return nil, err
	}

	if proof.d == nil {
		proof.d = &arith.Scalar{}
	}
	if m, err = proof.d.Unmarshal(m); err != nil {
		return nil, err
	}

	return m, nil
}
