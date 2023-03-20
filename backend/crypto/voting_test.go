package crypto

import (
	"crypto/rand"
	"testing"
)

func TestEncryptDecryptVote(t *testing.T) {
	tests := generateVotingTests()
	keyPair := generateKeyPair(t, rand.Reader)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			encryptedVote, _, err := tc.vote.Encrypt(rand.Reader, &keyPair.Pk)
			if err != nil {
				t.Fatal(err)
			}
			got, err := encryptedVote.Decrypt(&keyPair.Sk, tc.result)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.result {
				t.Fatalf("expected: %d, got: %d", tc.result, got)
			}
		})
	}
}

type votingTests map[string]struct {
	vote   Vote
	result int64
}

func generateVotingTests() votingTests {
	tests := votingTests{
		"yes": {vote: Yes, result: 1},
		"no":  {vote: No, result: 0},
		"42":  {vote: 42, result: 42},
	}
	return tests
}
