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

func TestMarshalUnmarshalProofSkKnowledge(t *testing.T) {
	proof := generateProofSkKnowledge(t)
	m, err := proof.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	newProof := new(ProofSkKnowledge)
	err = newProof.UnmarshalBinary(m)
	if err != nil {
		t.Fatal(err)
	}
	if !proof.b.Equal(&newProof.b) {
		t.Fatalf("b: got %s, expected %s", &newProof.b, &proof.b)
	}
	if !proof.d.Equal(&newProof.d) {
		t.Fatalf("d: got %s, expected %s", &newProof.d, &proof.d)
	}
}

func generateProofSkKnowledge(t *testing.T) *ProofSkKnowledge {
	keyPair := generateKeyPair(t, rand.Reader)
	proof, err := ProveSkKnowledge(rand.Reader, keyPair)
	if err != nil {
		t.Fatal(err)
	}
	return proof
}
