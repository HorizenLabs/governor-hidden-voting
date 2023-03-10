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

func TestUnmarshalOverDefaultProofSkKnowledge(t *testing.T) {
	// create default initialized proof
	proof := new(ProofSkKnowledge)
	// check that its fields are nil
	if proof.b != nil {
		t.Fatal("proofB.b != nil")
	}
	if proof.d != nil {
		t.Fatal("proofB.d != nil")
	}

	testUnmarshalProofSkKnowledge(t, proof)
}

func TestUnmarshalOverExistingProofSkKnowledge(t *testing.T) {
	// create a proof
	keyPair := generateKeyPair(t, rand.Reader)
	proof, err := ProveSkKnowledge(rand.Reader, keyPair)
	if err != nil {
		t.Fatal(err)
	}
	// check that its fields are not nil
	if proof.b == nil {
		t.Fatal("proofB.b == nil")
	}
	if proof.d == nil {
		t.Fatal("proofB.d == nil")
	}

	testUnmarshalProofSkKnowledge(t, proof)
}

func testUnmarshalProofSkKnowledge(t *testing.T, proofB *ProofSkKnowledge) {
	keyPair := generateKeyPair(t, rand.Reader)
	proofA, err := ProveSkKnowledge(rand.Reader, keyPair)
	if err != nil {
		t.Fatal(err)
	}

	// marshal proofA into mA, then unmarshal mA into proofB
	mA := proofA.Marshal()
	_, err = proofB.Unmarshal(mA)
	if err != nil {
		t.Fatal(err)
	}

	// check that proofA and proofB are equal
	if !proofA.b.Equal(proofB.b) {
		t.Fatalf("proofB.b = %s, expected %s", proofB.b, proofA.b)
	}
	if !proofA.d.Equal(proofB.d) {
		t.Fatalf("proofB.d = %s, expected %s", proofB.d, proofA.d)
	}
}
