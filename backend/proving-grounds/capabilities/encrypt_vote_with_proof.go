package capabilities

import (
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

type EncryptVoteWithProofCapability struct{}

func (EncryptVoteWithProofCapability) Id() string {
	return "772e6272-8c49-463f-b2d1-92088ae06da1"
}

func (EncryptVoteWithProofCapability) Name() string {
	return "encrypt_vote_with_proof"
}

func (EncryptVoteWithProofCapability) Description() string {
	return "encrypt a vote and generate proof of correct encryption"
}

func (EncryptVoteWithProofCapability) NumArguments() uint32 {
	return 2
}

func (c EncryptVoteWithProofCapability) Compute(in [][]byte) ([][]byte, error) {
	if len(in) != int(c.NumArguments()) {
		return nil, fmt.Errorf("%d arguments provided, but capability requires %d arguments", len(in), c.NumArguments())
	}
	vote := new(int64)
	pk := new(arith.CurvePoint)
	err := json.Unmarshal(in[0], vote)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(in[1], pk)
	if err != nil {
		return nil, err
	}

	encryptedVote, proof, err := crypto.EncryptVoteWithProof(rand.Reader, *vote, pk)
	if err != nil {
		return nil, err
	}

	encryptedVoteJSON, err := json.Marshal(encryptedVote)
	if err != nil {
		return nil, err
	}
	proofJSON, err := json.Marshal(proof)
	if err != nil {
		return nil, err
	}
	return [][]byte{encryptedVoteJSON, proofJSON}, nil
}
