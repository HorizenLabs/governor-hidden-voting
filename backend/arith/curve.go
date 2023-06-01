package arith

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

const NumBytesCurvePoint = 2 * 256 / 8

type CurvePoint struct {
	p bn256.G1
}

type CurvePointInternal struct {
	X *big.Int `json:"x"`
	Y *big.Int `json:"y"`
}

func newCurvePoint(p *bn256.G1) *CurvePoint {
	point := new(CurvePoint)
	point.p.Set(p)
	return point
}

func RandomCurvePoint(r io.Reader) (*Scalar, *CurvePoint, error) {
	for {
		k, p, err := bn256.RandomG1(r)
		if err != nil {
			return nil, nil, err
		}
		if k.Sign() != 0 {
			return NewScalar(k), newCurvePoint(p), err
		}
	}
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

func (a CurvePoint) MarshalBinary() ([]byte, error) {
	return a.p.Marshal(), nil
}

func (a *CurvePoint) UnmarshalBinary(m []byte) error {
	if len(m) != NumBytesCurvePoint {
		return fmt.Errorf("curve point should be represented with %d bytes", NumBytesCurvePoint)
	}
	_, err := a.p.Unmarshal(m)
	return err
}

func (e CurvePoint) MarshalJSON() ([]byte, error) {
	bytesE, err := e.MarshalBinary()
	if err != nil {
		return nil, err
	}
	a := CurvePointInternal{
		X: new(big.Int).SetBytes(bytesE[:NumBytesCurvePoint/2]),
		Y: new(big.Int).SetBytes(bytesE[NumBytesCurvePoint/2:]),
	}
	return json.Marshal(a)
}

func (e *CurvePoint) UnmarshalJSON(data []byte) error {
	var point CurvePointInternal
	err := json.Unmarshal(data, &point)
	if err != nil {
		return err
	}
	if len(point.X.Bytes()) > NumBytesCurvePoint/2 {
		return fmt.Errorf("curve point x-coordinate is too big")
	}
	if len(point.Y.Bytes()) > NumBytesCurvePoint/2 {
		return fmt.Errorf("curve point y-coordinate is too big")
	}
	var buf = make([]byte, NumBytesCurvePoint)
	point.X.FillBytes(buf[:NumBytesCurvePoint/2])
	point.Y.FillBytes(buf[NumBytesCurvePoint/2:])
	return e.UnmarshalBinary(buf)
}
