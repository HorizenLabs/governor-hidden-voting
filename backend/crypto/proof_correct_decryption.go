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
	u arith.CurvePoint
	v arith.CurvePoint
	s arith.Scalar
}

func ProveCorrectDecryption(
	reader io.Reader,
	tally *EncryptedTally,
	keyPair *KeyPair) (*ProofCorrectDecryption, error) {

	r, v, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}
	u := new(arith.CurvePoint).ScalarMult(&tally.votes.a, r)

	bytesPk, err := keyPair.Pk().MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesA, err := tally.votes.a.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesB, err := tally.votes.b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesU, err := u.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesV, err := v.MarshalBinary()
	if err != nil {
		return nil, err
	}
	c := arith.FiatShamirChallenge(
		bytesPk,
		bytesA,
		bytesB,
		bytesU,
		bytesV)

	s := new(arith.Scalar).Mul(c.Scalar(), keyPair.Sk())
	s = new(arith.Scalar).Add(r, s)
	proof := new(ProofCorrectDecryption)
	proof.u.Set(u)
	proof.v.Set(v)
	proof.s.Set(s)
	return proof, nil
}

func VerifyCorrectDecryption(
	proof *ProofCorrectDecryption,
	tally *EncryptedTally,
	result int64,
	pk *arith.CurvePoint) error {

	m, err := proof.MarshalBinary()
	if err != nil {
		return err
	}
	proof = new(ProofCorrectDecryption)
	err = proof.UnmarshalBinary(m)
	if err != nil {
		return err
	}

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA, err := tally.votes.a.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB, err := tally.votes.b.MarshalBinary()
	if err != nil {
		return err
	}
	bytesU, err := proof.u.MarshalBinary()
	if err != nil {
		return err
	}
	bytesV, err := proof.v.MarshalBinary()
	if err != nil {
		return err
	}
	c := arith.FiatShamirChallenge(
		bytesPk,
		bytesA,
		bytesB,
		bytesU,
		bytesV)
	d := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-result)))
	d = new(arith.CurvePoint).Add(d, &tally.votes.b)
	sA := new(arith.CurvePoint).ScalarMult(&tally.votes.a, &proof.s)
	cD := new(arith.CurvePoint).ScalarMult(d, c.Scalar())
	uPlusCD := new(arith.CurvePoint).Add(&proof.u, cD)
	if !sA.Equal(uPlusCD) {
		return errors.New("decryption proof verification failed, first check")
	}

	sG := new(arith.CurvePoint).ScalarBaseMult(&proof.s)
	cPk := new(arith.CurvePoint).ScalarMult(pk, c.Scalar())
	vPlusCPk := new(arith.CurvePoint).Add(&proof.v, cPk)
	if !sG.Equal(vPlusCPk) {
		return errors.New("decryption proof verification failed, second check")
	}
	return nil
}

func (proof *ProofCorrectDecryption) MarshalBinary() ([]byte, error) {
	bytesU, err := proof.u.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesV, err := proof.v.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesS, err := proof.s.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ret := make([]byte, 0, numBytesProofCorrectDecryption)
	ret = append(ret, bytesU...)
	ret = append(ret, bytesV...)
	ret = append(ret, bytesS...)
	return ret, nil
}

func (proof *ProofCorrectDecryption) UnmarshalBinary(m []byte) error {
	if len(m) < numBytesProofCorrectDecryption {
		return errors.New("ProofCorrectDecryption: not enough data")
	}
	var err error

	if err = proof.u.UnmarshalBinary(m); err != nil {
		return err
	}
	m = m[arith.NumBytesCurvePoint:]
	if err = proof.v.UnmarshalBinary(m); err != nil {
		return err
	}
	m = m[arith.NumBytesCurvePoint:]
	if err = proof.s.UnmarshalBinary(m); err != nil {
		return err
	}
	return nil
}
