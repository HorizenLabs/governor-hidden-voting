package capabilities

import (
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

type DecryptTallyWithProofCapability struct{}

func (DecryptTallyWithProofCapability) Id() string {
	return "c238d864-ae22-4db5-b3d1-83c41cb8b4dd"
}

func (DecryptTallyWithProofCapability) Name() string {
	return "decrypt_tally_with_proof"
}

func (DecryptTallyWithProofCapability) Description() string {
	return "decrypt an encrypted tally and generate proof of correct decryption"
}

func (DecryptTallyWithProofCapability) NumArguments() uint32 {
	return 3
}

func (c DecryptTallyWithProofCapability) Compute(in [][]byte) ([][]byte, error) {
	if len(in) != int(c.NumArguments()) {
		return nil, fmt.Errorf("%d arguments provided, but capability requires %d arguments", len(in), c.NumArguments())
	}
	tally := new(crypto.EncryptedVote)
	n := new(int64)
	keyPair := new(crypto.KeyPair)
	err := json.Unmarshal(in[0], tally)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(in[1], n)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(in[2], keyPair)
	if err != nil {
		return nil, err
	}

	decryptedVote, proof, err := crypto.DecryptTallyWithProof(rand.Reader, tally, *n, keyPair)
	if err != nil {
		return nil, err
	}

	decryptedVoteJSON, err := json.Marshal(decryptedVote)
	if err != nil {
		return nil, err
	}
	proofJSON, err := json.Marshal(proof)
	if err != nil {
		return nil, err
	}
	return [][]byte{decryptedVoteJSON, proofJSON}, nil
}
