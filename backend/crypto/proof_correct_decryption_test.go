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
			encryptedVote := generateEncryptedResult(
				t,
				rand.Reader,
				&keyPair.Pk,
				tc.numVoters,
				tc.numYes)
			proof, err := ProveCorrectDecryption(rand.Reader, encryptedVote, keyPair)
			if err != nil {
				t.Fatal(err)
			}
			err = VerifyCorrectDecryption(proof, encryptedVote, tc.numYes, &keyPair.Pk)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
