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
				&keyPair.Pk,
				tc.numVoters,
				tc.numYes)
			proof, err := ProveCorrectDecryption(rand.Reader, &encryptedTally.Votes, keyPair)
			if err != nil {
				t.Fatal(err)
			}
			err = VerifyCorrectDecryption(proof, &encryptedTally.Votes, tc.numYes, &keyPair.Pk)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
