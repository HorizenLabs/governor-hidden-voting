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
	a *arith.CurvePoint
	b *arith.CurvePoint
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
	if e.a == nil {
		e.a = &arith.CurvePoint{}
	}
	e.a.Set(v.a)
	if e.b == nil {
		e.b = &arith.CurvePoint{}
	}
	e.b.Set(v.b)
	return e
}

func (e *EncryptedVote) Add(a, b *EncryptedVote) *EncryptedVote {
	if e.a == nil {
		e.a = &arith.CurvePoint{}
	}
	e.a.Add(a.a, b.a)
	if e.b == nil {
		e.b = &arith.CurvePoint{}
	}
	e.b.Add(a.b, b.b)
	return e
}

func (vote Vote) Encrypt(reader io.Reader, pk *arith.CurvePoint) (*EncryptedVote, *arith.Scalar, error) {
	r, a, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, nil, err
	}
	b := new(arith.CurvePoint).ScalarMult(pk, r)
	b = new(arith.CurvePoint).Add(b, vote.curvePoint())
	return &EncryptedVote{a: a, b: b}, r, nil
}

func (vote *EncryptedVote) Decrypt(sk *arith.Scalar, n int64) (*int64, error) {
	// Decrypt the vote using baby-step giant-step algorithm
	target := new(arith.CurvePoint).ScalarMult(vote.a, new(arith.Scalar).Neg(sk))
	target = new(arith.CurvePoint).Add(target, vote.b)
	m := int64(math.Ceil(math.Sqrt(float64(n + 1))))
	gPow := make(map[string]int64)
	for j := int64(0); j < m; j++ {
		val := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(j)))
		gPow[string(val.Marshal())] = j
	}
	mNegG := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-m)))
	gamma := new(arith.CurvePoint).Set(target)
	for i := int64(0); i < m; i++ {
		j, ok := gPow[string(gamma.Marshal())]
		if ok {
			result := i*m + j
			return &result, nil
		}
		gamma.Add(gamma, mNegG)
	}
	return nil, errors.New("error during vote decryption")
}

func (vote *EncryptedVote) Marshal() []byte {
	ret := make([]byte, 0, arith.NumBytesCurvePoint*2)
	ret = append(ret, vote.a.Marshal()...)
	ret = append(ret, vote.b.Marshal()...)
	return ret
}

func (vote *EncryptedVote) Unmarshal(m []byte) ([]byte, error) {
	if len(m) < 2*arith.NumBytesCurvePoint {
		return nil, errors.New("EncryptedVote: not enough data")
	}
	v := new(EncryptedVote)
	m, err := v.a.Unmarshal(m)
	if err != nil {
		return nil, err
	}
	m, err = v.b.Unmarshal(m)
	if err != nil {
		return nil, err
	}
	*vote = *v
	return m, nil
}
