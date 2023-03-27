package arith

import (
	"crypto/rand"
	"encoding/json"
	"testing"
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
