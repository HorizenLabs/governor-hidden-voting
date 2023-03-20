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
