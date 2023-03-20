package arith

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

const NumBytesScalar = 256 / 8

type Scalar struct {
	val big.Int
}

func (e *Scalar) String() string {
	return e.val.String()
}

func NewScalar(val *big.Int) *Scalar {
	scalar := new(Scalar)
	if val.CmpAbs(bn256.Order) < 0 {
		if val.Sign() >= 0 {
			scalar.val.Set(val)
		} else {
			scalar.val.Set(new(big.Int).Add(val, bn256.Order))
		}
	} else {
		scalar.val.Set(new(big.Int).Mod(val, bn256.Order))
	}
	return scalar
}

func (e *Scalar) Set(a *Scalar) *Scalar {
	e.val.Set(&a.val)
	return e
}

func (e *Scalar) Add(a, b *Scalar) *Scalar {
	c := new(big.Int).Add(&a.val, &b.val)
	if c.Cmp(bn256.Order) > 0 {
		c.Sub(c, bn256.Order)
	}
	e.val.Set(c)
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

func (e *Scalar) MarshalBinary() ([]byte, error) {
	buf := make([]byte, NumBytesScalar)
	e.val.FillBytes(buf)
	return buf, nil
}

func (e *Scalar) UnmarshalBinary(m []byte) error {
	if len(m) != NumBytesScalar {
		return fmt.Errorf("scalar should be represented with %d bytes", NumBytesScalar)
	}
	ret := new(big.Int).SetBytes(m)
	if ret.CmpAbs(bn256.Order) >= 0 {
		return errors.New("scalar is over the field modulus")
	}
	e.val.Set(ret)
	return nil
}

func (e *Scalar) MarshalJSON() ([]byte, error) {
	bytesE, err := e.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return json.Marshal(bytesE)
}

func (e *Scalar) UnmarshalJSON(data []byte) error {
	var bytesE []byte
	err := json.Unmarshal(data, &bytesE)
	if err != nil {
		return err
	}
	return e.UnmarshalBinary(bytesE)
}
