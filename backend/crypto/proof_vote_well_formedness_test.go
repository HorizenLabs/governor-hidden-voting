package crypto

import (
	"crypto/rand"
	"testing"
)

func TestProveAndVerifyVoteWellFormedness(t *testing.T) {
	tests := generateVotingTests()
	keyPair := generateKeyPair(t, rand.Reader)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			encryptedVote, secret, err := tc.vote.Encrypt(rand.Reader, keyPair.Pk())
			if err != nil {
				t.Fatal(err)
			}
			proof, err := ProveVoteWellFormedness(rand.Reader, encryptedVote, tc.vote, secret, keyPair.Pk())
			if err != nil {
				t.Fatal(err)
			}
			err = VerifyVoteWellFormedness(proof, encryptedVote, keyPair.Pk())
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMarshalUnmarshalProofVoteWellFormedness(t *testing.T) {
	tests := generateVotingTests()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			proof := generateProofVoteWellFormedness(t, tc.vote)
			m, err := proof.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			newProof := new(ProofVoteWellFormedness)
			err = newProof.UnmarshalBinary(m)
			if err != nil {
				t.Fatal(err)
			}
			if !proof.a0.Equal(&newProof.a0) {
				t.Fatalf("a0: got %s, expected %s", &newProof.a0, &proof.a0)
			}
			if !proof.a1.Equal(&newProof.a1) {
				t.Fatalf("a1: got %s, expected %s", &newProof.a1, &proof.a1)
			}
			if !proof.b0.Equal(&newProof.b0) {
				t.Fatalf("b0: got %s, expected %s", &newProof.b0, &proof.b0)
			}
			if !proof.b1.Equal(&newProof.b1) {
				t.Fatalf("b1: got %s, expected %s", &newProof.b1, &proof.b1)
			}
			if !proof.r0.Equal(&newProof.r0) {
				t.Fatalf("r0: got %s, expected %s", &newProof.r0, &proof.r0)
			}
			if !proof.r1.Equal(&newProof.r1) {
				t.Fatalf("r1: got %s, expected %s", &newProof.r1, &proof.r1)
			}
			if !proof.c0.Equal(&newProof.c0) {
				t.Fatalf("c0: got %s, expected %s", &newProof.c0, &proof.c0)
			}
		})
	}
}

func generateProofVoteWellFormedness(t *testing.T, vote Vote) *ProofVoteWellFormedness {
	keyPair := generateKeyPair(t, rand.Reader)
	encryptedVote, secret, err := vote.Encrypt(rand.Reader, keyPair.Pk())
	if err != nil {
		t.Fatal(err)
	}
	proof, err := ProveVoteWellFormedness(rand.Reader, encryptedVote, Yes, secret, keyPair.Pk())
	if err != nil {
		t.Fatal(err)
	}
	return proof
}
