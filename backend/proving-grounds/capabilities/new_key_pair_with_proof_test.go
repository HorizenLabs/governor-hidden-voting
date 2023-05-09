package capabilities

import (
	"encoding/json"
	"testing"

	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

func TestNewKeyPairWithProof(t *testing.T) {
	tests := map[string]struct {
		in            [][]byte
		shouldSucceed bool
	}{
		"CorrectArgs": {in: [][]byte{}, shouldSucceed: true},
		"TooManyArgs": {in: [][]byte{[]byte("Hello world")}, shouldSucceed: false},
	}

	capability := &NewKeyPairWithProofCapability{}

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
				keyPair := new(crypto.KeyPair)
				proof := new(crypto.ProofSkKnowledge)
				if err := json.Unmarshal(out[0], keyPair); err != nil {
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
