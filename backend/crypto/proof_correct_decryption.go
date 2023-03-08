package crypto

import (
	"errors"
	"io"
	"math/big"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

const numBytesProofCorrectDecryption = 2*arith.NumBytesCurvePoint +
	arith.NumBytesScalar

type ProofCorrectDecryption struct {
	u *arith.CurvePoint
	v *arith.CurvePoint
	s *arith.Scalar
}

func ProveCorrectDecryption(
	reader io.Reader,
	tally *EncryptedTally,
	keyPair *KeyPair) (*ProofCorrectDecryption, error) {
	r, v, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}
	u := new(arith.CurvePoint).ScalarMult(tally.votes.a, r)
	c := arith.FiatShamirChallenge(
		keyPair.Pk().Marshal(),
		tally.votes.a.Marshal(),
		tally.votes.b.Marshal(),
		u.Marshal(),
		v.Marshal())
	s := new(arith.Scalar).Mul(c.Scalar(), keyPair.Sk())
	s = new(arith.Scalar).Add(r, s)
	return &ProofCorrectDecryption{u: u, v: v, s: s}, nil
}

func VerifyCorrectDecryption(
	proof *ProofCorrectDecryption,
	tally *EncryptedTally,
	result int64,
	pk *arith.CurvePoint) error {

	m := proof.Marshal()
	proof = new(ProofCorrectDecryption)
	proof.Unmarshal(m)

	c := arith.FiatShamirChallenge(
		pk.Marshal(),
		tally.votes.a.Marshal(),
		tally.votes.b.Marshal(),
		proof.u.Marshal(),
		proof.v.Marshal())

	d := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-result)))
	d = new(arith.CurvePoint).Add(d, tally.votes.b)
	sA := new(arith.CurvePoint).ScalarMult(tally.votes.a, proof.s)
	cD := new(arith.CurvePoint).ScalarMult(d, c.Scalar())
	uPlusCD := new(arith.CurvePoint).Add(proof.u, cD)
	if !sA.Equal(uPlusCD) {
		return errors.New("decryption proof verification failed, first check")
	}

	sG := new(arith.CurvePoint).ScalarBaseMult(proof.s)
	cPk := new(arith.CurvePoint).ScalarMult(pk, c.Scalar())
	vPlusCPk := new(arith.CurvePoint).Add(proof.v, cPk)
	if !sG.Equal(vPlusCPk) {
		return errors.New("decryption proof verification failed, second check")
	}
	return nil
}

func (proof *ProofCorrectDecryption) Marshal() []byte {
	ret := make([]byte, 0, numBytesProofCorrectDecryption)
	ret = append(ret, proof.u.Marshal()...)
	ret = append(ret, proof.v.Marshal()...)
	ret = append(ret, proof.s.Marshal()...)
	return ret
}

func (proof *ProofCorrectDecryption) Unmarshal(m []byte) ([]byte, error) {
	if len(m) < numBytesProofCorrectDecryption {
		return nil, errors.New("ProofCorrectDecryption: not enough data")
	}
	var err error

	if proof.u == nil {
		proof.u = &arith.CurvePoint{}
	}
	if m, err = proof.u.Unmarshal(m); err != nil {
		return nil, err
	}

	if proof.v == nil {
		proof.v = &arith.CurvePoint{}
	}
	if m, err = proof.v.Unmarshal(m); err != nil {
		return nil, err
	}

	if proof.s == nil {
		proof.s = &arith.Scalar{}
	}
	if m, err = proof.s.Unmarshal(m); err != nil {
		return nil, err
	}
	return m, nil
}
