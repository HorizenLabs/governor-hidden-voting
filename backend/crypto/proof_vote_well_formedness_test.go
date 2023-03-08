package crypto

import (
	"crypto/rand"
	"testing"
)

func TestProveAndVerifyVoteWellFormedness(t *testing.T) {
	tests := map[string]struct {
		vote Vote
	}{
		"yes": {vote: Yes},
		"no":  {vote: No},
	}
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
