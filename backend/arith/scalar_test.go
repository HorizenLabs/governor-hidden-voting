package arith

import (
	"encoding/json"
	"math/big"
	"testing"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

func TestNewScalarBN256Order(t *testing.T) {
	a := NewScalar(big.NewInt(0))
	b := NewScalar(bn256.Order)
	if !a.Equal(b) {
		t.Fatal("NewScalar() is not wrapping around properly")
	}
}

func TestNewScalarMinusOne(t *testing.T) {
	a := NewScalar(big.NewInt(-1))
	n := new(big.Int).Set(bn256.Order)
	n = new(big.Int).Sub(n, big.NewInt(1))
	b := NewScalar(n)
	if !a.Equal(b) {
		t.Fatal("NewScalar() is not wrapping around properly")
	}
}

func TestUnmarshalInvalidScalar(t *testing.T) {
	tests := map[string]struct {
		m []byte
	}{
		"too short":    {m: make([]byte, NumBytesScalar-1)},
		"too long":     {m: make([]byte, NumBytesScalar+1)},
		"over modulus": {m: bn256.Order.Bytes()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			a := new(Scalar)
			err := a.UnmarshalBinary(tc.m)
			if err == nil {
				t.Fatalf("should be impossible to unmarshal a scalar %s", name)
			}
		})
	}

}

func TestMarshalUnmarshalScalar(t *testing.T) {
	tests := map[string]struct {
		n *big.Int
	}{
		"zero":      {n: big.NewInt(0)},
		"one":       {n: big.NewInt(1)},
		"random":    {n: big.NewInt(1234)},
		"minus one": {n: big.NewInt(-1)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			a := NewScalar(tc.n)
			m, err := a.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			aUnmarshaled := new(Scalar)
			err = aUnmarshaled.UnmarshalBinary(m)
			if err != nil {
				t.Fatal(err)
			}
			if !aUnmarshaled.Equal(a) {
				t.Fatal("unmarshaled a is different from a")
			}
		})
	}
}

func TestMarshalUnmarshalJSONScalar(t *testing.T) {
	tests := map[string]struct {
		n *big.Int
	}{
		"zero":      {n: big.NewInt(0)},
		"one":       {n: big.NewInt(1)},
		"random":    {n: big.NewInt(1234)},
		"minus one": {n: big.NewInt(-1)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			a := NewScalar(tc.n)
			m, err := json.Marshal(a)
			if err != nil {
				t.Fatal(err)
			}
			aUnmarshaled := new(Scalar)
			err = json.Unmarshal(m, aUnmarshaled)
			if err != nil {
				t.Fatal(err)
			}
			if !aUnmarshaled.Equal(a) {
				t.Fatal("unmarshaled a is different from a")
			}
		})
	}
}

func TestMarshalledScalarSize(t *testing.T) {
	tests := map[string]struct {
		n *big.Int
	}{
		"zero":      {n: big.NewInt(0)},
		"one":       {n: big.NewInt(1)},
		"random":    {n: big.NewInt(1234)},
		"minus one": {n: big.NewInt(-1)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			a := NewScalar(tc.n)
			m, err := a.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			if len(m) != NumBytesScalar {
				t.Fatalf("marshalled scalar byte length is %d, should be %d", len(m), NumBytesScalar)
			}
		})
	}
}
