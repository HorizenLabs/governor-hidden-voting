package crypto

import (
	"crypto/rand"
	"testing"
)

func TestProveAndVerifyVoteWellFormedness(t *testing.T) {
	tests := generateVoteWellFormednessTests()
	keyPair := generateKeyPair(t, rand.Reader)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			encryptedVote, secret, err := tc.vote.Encrypt(rand.Reader, &keyPair.Pk)
			if err != nil {
				t.Fatal(err)
			}
			proof, err := ProveVoteWellFormedness(rand.Reader, encryptedVote, tc.vote, secret, &keyPair.Pk)
			if tc.isProvable && err != nil {
				t.Fatal(err)
			}
			if !tc.isProvable && err == nil {
				t.Fatal("generated a proof of vote well formedness of invalid vote")
			}
			err = VerifyVoteWellFormedness(proof, encryptedVote, &keyPair.Pk)
			if tc.isProvable && err != nil {
				t.Fatal(err)
			}
			if !tc.isProvable && err == nil {
				t.Fatal("correctly verified a proof of vote well formedness of invalid vote")
			}
		})
	}
}

func TestVerifyVoteWellFormednessWithDifferentVote(t *testing.T) {
	tests := generateVoteWellFormednessTests()
	keyPair := generateKeyPair(t, rand.Reader)

	for name, tc := range tests {
		if tc.isProvable {
			t.Run(name, func(t *testing.T) {
				encryptedVote, secret, err := tc.vote.Encrypt(rand.Reader, &keyPair.Pk)
				if err != nil {
					t.Fatal(err)
				}
				proof, err := ProveVoteWellFormedness(rand.Reader, encryptedVote, tc.vote, secret, &keyPair.Pk)
				if err != nil {
					t.Fatal(err)
				}
				newVote, _, err := tc.vote.Encrypt(rand.Reader, &keyPair.Pk)
				if err != nil {
					t.Fatal(err)
				}
				err = VerifyVoteWellFormedness(proof, newVote, &keyPair.Pk)
				if err == nil {
					t.Fatal("correctly verified a proof of vote well formedness for an encrypted vote different from the one used to generate it")
				}
			})
		}
	}
}

type voteWellFormednessTests map[string]struct {
	vote       Vote
	isProvable bool
}

func generateVoteWellFormednessTests() voteWellFormednessTests {
	tests := voteWellFormednessTests{
		"yes": {vote: Yes, isProvable: true},
		"no":  {vote: No, isProvable: true},
		"42":  {vote: 42, isProvable: false},
	}
	return tests
}
