package arith

import (
	"crypto/rand"
	"encoding/json"
	"testing"
)

func TestUnmarshalInvalidChallenge(t *testing.T) {
	tests := map[string]struct {
		m []byte
	}{
		"too short": {m: make([]byte, NumBytesChallenge-1)},
		"too long":  {m: make([]byte, NumBytesChallenge+1)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			a := new(Scalar)
			err := a.UnmarshalBinary(tc.m)
			if err == nil {
				t.Fatalf("should be impossible to unmarshal a challenge %s", name)
			}
		})
	}
}

func TestMarshalUnmarshalChallenge(t *testing.T) {
	want, err := RandomChallenge(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	m, err := want.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	got := new(Challenge)
	err = got.UnmarshalBinary(m)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("want: %s, got: %s", want, got)
	}
}

func TestMarshalUnmarshalJSONChallenge(t *testing.T) {
	want, err := RandomChallenge(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	m, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got := new(Challenge)
	err = json.Unmarshal(m, got)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("want: %s, got: %s", want, got)
	}
}
