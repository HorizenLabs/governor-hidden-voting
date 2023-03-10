package arith

import (
	"errors"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

const NumBytesScalar = 256 / 8

type Scalar struct {
	val big.Int
}

func (c *Scalar) String() string {
	return c.val.String()
}

func NewScalar(val *big.Int) *Scalar {
	return &Scalar{*new(big.Int).Mod(val, bn256.Order)}
}

func (e *Scalar) Add(a, b *Scalar) *Scalar {
	e.val.Add(&a.val, &b.val)
	e.val.Mod(&e.val, bn256.Order)
	return e
}

func (e *Scalar) Mul(a, b *Scalar) *Scalar {
	e.val.Mul(&a.val, &b.val)
	e.val.Mod(&e.val, bn256.Order)
	return e
}

func (e *Scalar) Neg(a *Scalar) *Scalar {
	e.val.Set(new(big.Int).Sub(bn256.Order, &a.val))
	return e
}

func (a *Scalar) Equal(b *Scalar) bool {
	return a.val.Cmp(&b.val) == 0
}

func (e *Scalar) Marshal() []byte {
	buf := make([]byte, NumBytesScalar)
	e.val.FillBytes(buf)
	return buf
}

func (e *Scalar) Unmarshal(m []byte) ([]byte, error) {
	if len(m) < NumBytesScalar {
		return nil, errors.New("message too short")
	}
	ret := new(big.Int).SetBytes(m[:NumBytesScalar])
	if ret.CmpAbs(bn256.Order) >= 0 {
		return nil, errors.New("scalar is over the field modulus")
	}
	e.val.Set(ret)
	return m[NumBytesScalar:], nil
}
