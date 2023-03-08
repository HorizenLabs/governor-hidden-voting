package crypto

import (
	"crypto/rand"
	"testing"
)

func TestProveAndVerifySkKnowledge(t *testing.T) {
	keyPair := generateKeyPair(t, rand.Reader)
	proof, err := ProveSkKnowledge(rand.Reader, keyPair)
	if err != nil {
		t.Fatal(err)
	}
	err = VerifySkKnowledge(proof, keyPair.Pk())
	if err != nil {
		t.Fatal(err)
	}
}
