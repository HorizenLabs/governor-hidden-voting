package capabilities

import (
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

func TestDecryptTallyWithProof(t *testing.T) {
	correctInputs := decryptTallyWithProofInputsFixture(t)

	tests := map[string]struct {
		in            [][]byte
		shouldSucceed bool
	}{
		"CorrectArgs": {in: correctInputs, shouldSucceed: true},
		"TooFewArgs":  {in: correctInputs[0:2], shouldSucceed: false},
		"TooManyArgs": {in: [][]byte{correctInputs[0], correctInputs[1], correctInputs[2], []byte("Hello world")}, shouldSucceed: false},
		"WrongArg0":   {in: [][]byte{[]byte("Hello world"), correctInputs[1], correctInputs[2]}, shouldSucceed: false},
		"WrongArg1":   {in: [][]byte{correctInputs[0], []byte("Hello world"), correctInputs[2]}, shouldSucceed: false},
		"WrongArg2":   {in: [][]byte{correctInputs[0], correctInputs[1], []byte("Hello world")}, shouldSucceed: false},
	}

	capability := &DecryptTallyWithProofCapability{}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			out, err := capability.Compute(tc.in)
			if tc.shouldSucceed {
				if err != nil {
					t.Fatal(err)
				}
				if len(out) != 2 {
					t.Fatalf("result should contain 2 outputs")
				}
				decryptedVote := new(int64)
				proof := new(crypto.ProofCorrectDecryption)
				if err := json.Unmarshal(out[0], decryptedVote); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(out[1], proof); err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil {
					t.Fatalf("compute should fail")
				}
			}
		})
	}
}

func decryptTallyWithProofInputsFixture(t *testing.T) [][]byte {
	keyPair, err := crypto.NewKeyPair(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	keyPairJSON, err := json.Marshal(keyPair)
	if err != nil {
		t.Fatal(err)
	}
	vote := int64(1)
	encryptedVote, _, err := crypto.EncryptVoteWithProof(rand.Reader, vote, &keyPair.Pk)
	if err != nil {
		t.Fatal(err)
	}
	encryptedVoteJSON, err := json.Marshal(encryptedVote)
	if err != nil {
		t.Fatal(err)
	}
	nJSON, err := json.Marshal(int64(1))
	if err != nil {
		t.Fatal(err)
	}
	return [][]byte{encryptedVoteJSON, nJSON, keyPairJSON}
}
