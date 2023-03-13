package arith

import (
	"bytes"
	"io"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

const NumBytesCurvePoint = 2 * 256 / 8

type CurvePoint struct {
	p bn256.G1
}

func newCurvePoint(p *bn256.G1) *CurvePoint {
	point := new(CurvePoint)
	point.p.Set(p)
	return point
}

func RandomCurvePoint(r io.Reader) (*Scalar, *CurvePoint, error) {
	k, p, err := bn256.RandomG1(r)
	return NewScalar(k), newCurvePoint(p), err
}

func (a *CurvePoint) String() string {
	return a.p.String()
}

func (e *CurvePoint) Set(a *CurvePoint) *CurvePoint {
	e.p.Set(&a.p)
	return e
}

func (e *CurvePoint) ScalarBaseMult(k *Scalar) *CurvePoint {
	e.p.Set(new(bn256.G1).ScalarBaseMult(&k.val))
	return e
}

func (e *CurvePoint) ScalarMult(a *CurvePoint, k *Scalar) *CurvePoint {
	e.p.Set(new(bn256.G1).ScalarMult(&a.p, &k.val))
	return e
}

func (e *CurvePoint) Add(a, b *CurvePoint) *CurvePoint {
	e.p.Set(new(bn256.G1).Add(&a.p, &b.p))
	return e
}

func (e *CurvePoint) Neg(a *CurvePoint) *CurvePoint {
	e.p.Set(new(bn256.G1).Neg(&a.p))
	return e
}

func (a *CurvePoint) Equal(b *CurvePoint) bool {
	return bytes.Equal(a.p.Marshal(), b.p.Marshal())
}

func (a *CurvePoint) MarshalBinary() ([]byte, error) {
	return a.p.Marshal(), nil
}

func (a *CurvePoint) UnmarshalBinary(m []byte) error {
	_, err := a.p.Unmarshal(m)
	return err
}
