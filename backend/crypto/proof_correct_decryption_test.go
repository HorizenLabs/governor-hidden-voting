package crypto

import (
	"crypto/rand"
	"testing"
)

func TestProveAndVerifyCorrectDecryption(t *testing.T) {
	tests := generateTallyingTests()
	keyPair := generateKeyPair(t, rand.Reader)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			encryptedTally := generateEncryptedTally(
				t,
				rand.Reader,
				keyPair.Pk(),
				tc.numVoters,
				tc.numYes)
			proof, err := ProveCorrectDecryption(rand.Reader, encryptedTally, keyPair)
			if err != nil {
				t.Fatal(err)
			}
			err = VerifyCorrectDecryption(proof, encryptedTally, tc.numYes, keyPair.Pk())
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMarshalUnmarshalProofCorrectDecryption(t *testing.T) {
	proof := generateProofCorrectDecryption(t)
	m, err := proof.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	newProof := new(ProofCorrectDecryption)
	err = newProof.UnmarshalBinary(m)
	if err != nil {
		t.Fatal(err)
	}
	if !proof.u.Equal(&newProof.u) {
		t.Fatalf("s: got %s, expected %s", &newProof.u, &proof.u)
	}
	if !proof.v.Equal(&newProof.v) {
		t.Fatalf("v: got %s, expected %s", &newProof.v, &proof.v)
	}
	if !proof.s.Equal(&newProof.s) {
		t.Fatalf("s: got %s, expected %s", &newProof.s, &proof.s)
	}
}

func generateProofCorrectDecryption(t *testing.T) *ProofCorrectDecryption {
	keyPair := generateKeyPair(t, rand.Reader)
	encryptedTally := generateEncryptedTally(
		t,
		rand.Reader,
		keyPair.Pk(),
		1,
		1)
	proof, err := ProveCorrectDecryption(rand.Reader, encryptedTally, keyPair)
	if err != nil {
		t.Fatal(err)
	}
	return proof
}
