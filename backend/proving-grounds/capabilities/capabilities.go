package capabilities

import (
	"crypto/rand"
	"encoding/json"
	"sort"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
	"golang.org/x/exp/maps"
)

const NEW_KEY_PAIR_WITH_PROOF_ID = "924e949e-0f98-42c6-a4ff-bb2b1c8693e0"
const ENCRYPT_VOTE_WITH_PROOF_ID = "772e6272-8c49-463f-b2d1-92088ae06da1"
const DECRYPT_TALLY_WITH_PROOF_ID = "c238d864-ae22-4db5-b3d1-83c41cb8b4dd"

var Capabilities = map[string]*cs.Capability{
	NEW_KEY_PAIR_WITH_PROOF_ID: {
		Id:           NEW_KEY_PAIR_WITH_PROOF_ID,
		Name:         "new_key_pair_with_proof",
		Description:  "generate ElGamal keypair and proof of sk knowledge",
		NumArguments: 0,
	},
	ENCRYPT_VOTE_WITH_PROOF_ID: {
		Id:           ENCRYPT_VOTE_WITH_PROOF_ID,
		Name:         "encrypt_vote_with_proof",
		Description:  "encrypt a vote and generate proof of correct encryption",
		NumArguments: 2,
	},
	DECRYPT_TALLY_WITH_PROOF_ID: {
		Id:           DECRYPT_TALLY_WITH_PROOF_ID,
		Name:         "decrypt_tally_with_proof",
		Description:  "decrypt an encrypted tally and generate proof of correct decryption",
		NumArguments: 3,
	},
}

func ListCapabilities() []*cs.Capability {
	keys := maps.Keys(Capabilities)
	sort.Strings(keys)
	var capabilities []*cs.Capability
	for _, k := range keys {
		capabilities = append(capabilities, Capabilities[k])
	}
	return capabilities
}

var Handlers = map[string]func(args [][]byte) ([][]byte, error){
	NEW_KEY_PAIR_WITH_PROOF_ID:  func(args [][]byte) ([][]byte, error) { return HandleNewKeyPairWithProof(args) },
	ENCRYPT_VOTE_WITH_PROOF_ID:  func(args [][]byte) ([][]byte, error) { return HandleEncryptVoteWithProof(args) },
	DECRYPT_TALLY_WITH_PROOF_ID: func(args [][]byte) ([][]byte, error) { return HandleDecryptTallyWithProof(args) },
}

func HandleNewKeyPairWithProof(args [][]byte) ([][]byte, error) {
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

func HandleEncryptVoteWithProof(args [][]byte) ([][]byte, error) {
	vote := new(int64)
	pk := new(arith.CurvePoint)
	err := json.Unmarshal(args[0], vote)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(args[1], pk)
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

func HandleDecryptTallyWithProof(args [][]byte) ([][]byte, error) {
	tally := new(crypto.EncryptedVote)
	n := new(int64)
	keyPair := new(crypto.KeyPair)
	err := json.Unmarshal(args[0], tally)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(args[1], n)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(args[2], keyPair)
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
