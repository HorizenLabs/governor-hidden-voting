package capabilities

import (
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

func TestEncryptVoteWithProof(t *testing.T) {
	correctInputs := encryptVoteWithProofInputsFixture(t)

	tests := map[string]struct {
		in            [][]byte
		shouldSucceed bool
	}{
		"CorrectArgs": {in: correctInputs, shouldSucceed: true},
		"TooFewArgs":  {in: [][]byte{correctInputs[0]}, shouldSucceed: false},
		"TooManyArgs": {in: [][]byte{correctInputs[0], correctInputs[1], []byte("Hello world")}, shouldSucceed: false},
		"WrongArg0":   {in: [][]byte{[]byte("Hello world"), correctInputs[1]}, shouldSucceed: false},
		"WrongArg1":   {in: [][]byte{correctInputs[0], []byte("Hello world")}, shouldSucceed: false},
	}

	capability := &EncryptVoteWithProofCapability{}

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
				encryptedVote := new(crypto.EncryptedVote)
				proof := new(crypto.ProofVoteWellFormedness)
				if err := json.Unmarshal(out[0], encryptedVote); err != nil {
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

func encryptVoteWithProofInputsFixture(t *testing.T) [][]byte {
	voteJSON, err := json.Marshal(int64(1))
	if err != nil {
		t.Fatal(err)
	}
	keyPair, err := crypto.NewKeyPair(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pkJSON, err := json.Marshal(keyPair.Pk)
	if err != nil {
		t.Fatal(err)
	}
	return [][]byte{voteJSON, pkJSON}
}
