package crypto

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

func TestNewKeyPair(t *testing.T) {
	keyPair := generateKeyPair(t, rand.Reader)
	pk := new(arith.CurvePoint).ScalarBaseMult(keyPair.Sk())
	if !pk.Equal(keyPair.Pk()) {
		t.Fatal(`keyPair.Pk() != G1 * keyPair.Sk()`)
	}
}

func generateKeyPair(t *testing.T, r io.Reader) *KeyPair {
	keyPair, err := NewKeyPair(r)
	if err != nil {
		t.Fatal(err)
	}
	return keyPair
}
