package capabilities

import (
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

type NewKeyPairWithProofCapability struct{}

func (NewKeyPairWithProofCapability) Id() string {
	return "924e949e-0f98-42c6-a4ff-bb2b1c8693e0"
}

func (NewKeyPairWithProofCapability) Name() string {
	return "new_key_pair_with_proof"
}

func (NewKeyPairWithProofCapability) Description() string {
	return "generate ElGamal keypair and proof of sk knowledge"
}

func (NewKeyPairWithProofCapability) NumArguments() uint32 {
	return 0
}

func (c NewKeyPairWithProofCapability) Compute(in [][]byte) ([][]byte, error) {
	if len(in) != int(c.NumArguments()) {
		return nil, fmt.Errorf("%d arguments provided, but capability requires %d arguments", len(in), c.NumArguments())
	}
	keypair, proof, err := crypto.NewKeyPairWithProof(rand.Reader)
	if err != nil {
		return nil, err
	}
	keyPairJSON, err := json.Marshal(keypair)
	if err != nil {
		return nil, err
	}
	proofJSON, err := json.Marshal(proof)
	if err != nil {
		return nil, err
	}
	return [][]byte{keyPairJSON, proofJSON}, nil
}
