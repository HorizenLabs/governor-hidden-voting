package crypto

import (
	"crypto/rand"
	"testing"
)

func TestEncryptDecryptVote(t *testing.T) {
	tests := map[string]struct {
		vote Vote
		want int64
	}{
		"yes": {vote: Yes, want: 1},
		"no":  {vote: No, want: 0},
	}
	keyPair := generateKeyPair(t, rand.Reader)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			encryptedVote, _, err := tc.vote.Encrypt(rand.Reader, keyPair.Pk())
			if err != nil {
				t.Fatal(err)
			}
			got, err := encryptedVote.Decrypt(keyPair.Sk(), 1)
			if err != nil {
				t.Fatal(err)
			}
			if *got != tc.want {
				t.Fatalf("expected: %d, got: %d", tc.want, *got)
			}
		})
	}
}
