package arith

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"testing"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

func TestUnmarshalCurvePoint(t *testing.T) {
	_, p, err := RandomCurvePoint(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	mCorrect, err := p.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	mOutsideX := make([]byte, len(mCorrect))
	copy(mOutsideX, mCorrect)
	mOutsideX[16] = mOutsideX[16] + 1

	mOutsideY := make([]byte, len(mCorrect))
	copy(mOutsideY, mCorrect)
	mOutsideY[48] = mOutsideY[48] + 1

	tests := map[string]struct {
		m          []byte
		shouldPass bool
	}{
		"correct":               {m: mCorrect, shouldPass: true},
		"too short":             {m: make([]byte, NumBytesCurvePoint-1), shouldPass: false},
		"too long":              {m: make([]byte, NumBytesCurvePoint+1), shouldPass: false},
		"x coord outside curve": {m: mOutsideX, shouldPass: false},
		"y coord outside curve": {m: mOutsideY, shouldPass: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			a := new(CurvePoint)
			err := a.UnmarshalBinary(tc.m)
			if tc.shouldPass && err != nil {
				t.Fatal("cannot unmarshal valid curve point")
			}
			if !tc.shouldPass && err == nil {
				t.Fatalf("successfully marshaled invalid curve point: %s", name)
			}
		})
	}
}

func TestMarshalUnmarshalCurvePoint(t *testing.T) {
	_, want, err := RandomCurvePoint(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	m, err := want.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	got := new(CurvePoint)
	err = got.UnmarshalBinary(m)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("want: %s, got: %s", want, got)
	}
}

func TestMarshalUnmarshalJSONCurvePoint(t *testing.T) {
	_, want, err := RandomCurvePoint(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	m, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got := new(CurvePoint)
	err = json.Unmarshal(m, got)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("want: %s, got: %s", want, got)
	}
}

func TestCoordinateValues(t *testing.T) {
	tests := map[string]struct {
		p *CurvePoint
		x *big.Int
		y *big.Int
	}{
		"g": {
			p: new(CurvePoint).ScalarBaseMult(NewScalar(big.NewInt(1))),
			x: big.NewInt(1),
			y: big.NewInt(2),
		},
		"gNeg": {
			p: new(CurvePoint).ScalarBaseMult(NewScalar(big.NewInt(-1))),
			x: big.NewInt(1),
			y: new(big.Int).Sub(bn256.P, big.NewInt(2)),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m, err := json.Marshal(tc.p)
			if err != nil {
				t.Fatal(err)
			}
			var point CurvePointInternal
			err = json.Unmarshal(m, &point)
			if err != nil {
				t.Fatal(err)
			}
			x := new(big.Int).SetBytes(point.X)
			y := new(big.Int).SetBytes(point.Y)

			if x.Cmp(tc.x) != 0 {
				t.Fatalf("expected x == %s, got %s", tc.x, x)
			}
			if y.Cmp(tc.y) != 0 {
				t.Fatalf("expected y == %s, got %s", tc.y, y)
			}
		})
	}
}
