package crypto

import (
	"errors"
	"io"
	"math"
	"math/big"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

type Vote int

const (
	No Vote = iota
	Yes
)

type EncryptedVote struct {
	A arith.CurvePoint
	B arith.CurvePoint
}

func (vote Vote) curvePoint() *arith.CurvePoint {
	var scalar *arith.Scalar
	switch vote {
	case Yes:
		scalar = arith.NewScalar(big.NewInt(1))
	case No:
		scalar = arith.NewScalar(big.NewInt(0))
	}
	return new(arith.CurvePoint).ScalarBaseMult(scalar)
}

func (e *EncryptedVote) Set(v *EncryptedVote) *EncryptedVote {
	e.A.Set(&v.A)
	e.B.Set(&v.B)
	return e
}

func (e *EncryptedVote) Add(a, b *EncryptedVote) *EncryptedVote {
	e.A.Add(&a.A, &b.A)
	e.B.Add(&a.B, &b.B)
	return e
}

func (vote Vote) Encrypt(reader io.Reader, pk *arith.CurvePoint) (*EncryptedVote, *arith.Scalar, error) {
	r, a, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, nil, err
	}
	b := new(arith.CurvePoint).ScalarMult(pk, r)
	b = new(arith.CurvePoint).Add(b, vote.curvePoint())
	encryptedVote := new(EncryptedVote)
	encryptedVote.A.Set(a)
	encryptedVote.B.Set(b)
	return encryptedVote, r, nil
}

func (vote *EncryptedVote) Decrypt(sk *arith.Scalar, n int64) (int64, error) {
	// Decrypt the vote using baby-step giant-step algorithm
	target := new(arith.CurvePoint).ScalarMult(&vote.A, new(arith.Scalar).Neg(sk))
	target = new(arith.CurvePoint).Add(target, &vote.B)
	m := int64(math.Ceil(math.Sqrt(float64(n + 1))))
	gPow := make(map[string]int64)
	for j := int64(0); j < m; j++ {
		val := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(j)))
		valBinary, err := val.MarshalBinary()
		if err != nil {
			return 0, err
		}
		gPow[string(valBinary)] = j
	}
	mNegG := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-m)))
	gamma := new(arith.CurvePoint).Set(target)
	for i := int64(0); i < m; i++ {
		gammaBinary, err := gamma.MarshalBinary()
		if err != nil {
			return 0, err
		}
		j, ok := gPow[string(gammaBinary)]
		if ok {
			result := i*m + j
			return result, nil
		}
		gamma.Add(gamma, mNegG)
	}
	return 0, errors.New("error during vote decryption")
}
