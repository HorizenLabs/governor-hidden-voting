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

func TestMarshalUnmarshalKeyPair(t *testing.T) {
	keyPair := generateKeyPair(t, rand.Reader)
	m, err := keyPair.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	newKeyPair := new(KeyPair)
	err = newKeyPair.UnmarshalBinary(m)
	if err != nil {
		t.Fatal(err)
	}
	if !keyPair.pk.Equal(&newKeyPair.pk) {
		t.Fatalf("pk: got %s, expected %s", &newKeyPair.pk, &keyPair.pk)
	}
	if !keyPair.sk.Equal(&newKeyPair.sk) {
		t.Fatalf("sk: got %s, expected %s", &newKeyPair.sk, &keyPair.sk)
	}
}

func generateKeyPair(t *testing.T, r io.Reader) *KeyPair {
	keyPair, err := NewKeyPair(r)
	if err != nil {
		t.Fatal(err)
	}
	return keyPair
}
