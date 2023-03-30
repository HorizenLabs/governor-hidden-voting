// Generate test data for smart contract tests
package main

import (
	"crypto/rand"
	"encoding/json"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

type TestData struct {
	PkA                             arith.CurvePoint
	ProofSkKnowledgeA               crypto.ProofSkKnowledge
	PkB                             arith.CurvePoint
	ProofSkKnowledgeB               crypto.ProofSkKnowledge
	EncryptedVotesValid             []crypto.EncryptedVote
	ProofsVoteWellFormednessValid   []crypto.ProofVoteWellFormedness
	EncryptedVotesInvalid           []crypto.EncryptedVote
	ProofsVoteWellFormednessInvalid []crypto.ProofVoteWellFormedness
	ProofCorrectDecryptionValid     crypto.ProofCorrectDecryption
	ProofCorrectDecryptionInvalid   crypto.ProofCorrectDecryption
	Result                          int64
}

func main() {
	testData := new(TestData)

	keyPairA, proofSkKnowledgeA, err := crypto.NewKeyPairWithProof(rand.Reader)
	if err != nil {
		panic(err)
	}
	pkA := &keyPairA.Pk

	keyPairB, proofSkKnowledgeB, err := crypto.NewKeyPairWithProof(rand.Reader)
	if err != nil {
		panic(err)
	}
	pkB := &keyPairB.Pk

	encryptedTally := crypto.NewEncryptedVote()

	votes := []crypto.Vote{crypto.Yes, crypto.No, crypto.Yes, crypto.Yes, crypto.No}
	for _, vote := range votes {
		encryptedVote, proof, err := crypto.EncryptVoteWithProof(rand.Reader, int64(vote), pkA)
		if err != nil {
			panic(err)
		}
		encryptedTally.Add(encryptedTally, encryptedVote)
		testData.EncryptedVotesValid = append(testData.EncryptedVotesValid, *encryptedVote)
		testData.ProofsVoteWellFormednessValid = append(testData.ProofsVoteWellFormednessValid, *proof)
	}

	votesInvalid := []crypto.Vote{crypto.Yes, crypto.No}
	for _, vote := range votesInvalid {
		encryptedVote, proof, err := crypto.EncryptVoteWithProof(rand.Reader, int64(vote), pkB)
		if err != nil {
			panic(err)
		}
		testData.EncryptedVotesInvalid = append(testData.EncryptedVotesInvalid, *encryptedVote)
		testData.ProofsVoteWellFormednessInvalid = append(testData.ProofsVoteWellFormednessInvalid, *proof)
	}

	result, err := encryptedTally.Decrypt(&keyPairA.Sk, int64(len(votes)))
	if err != nil {
		panic(err)
	}

	proofCorrectDecryptionValid, err := crypto.ProveCorrectDecryption(rand.Reader, encryptedTally, keyPairA)
	if err != nil {
		panic(err)
	}

	err = crypto.VerifyCorrectDecryption(proofCorrectDecryptionValid, encryptedTally, result, pkA)
	if err != nil {
		panic(err)
	}

	additionalVote, _, err := crypto.EncryptVoteWithProof(rand.Reader, int64(crypto.No), pkA)
	if err != nil {
		panic(err)
	}
	encryptedTally.Add(encryptedTally, additionalVote)
	proofCorrectDecryptionInvalid, err := crypto.ProveCorrectDecryption(rand.Reader, encryptedTally, keyPairA)
	if err != nil {
		panic(err)
	}

	testData.PkA.Set(pkA)
	testData.ProofSkKnowledgeA.Set(proofSkKnowledgeA)
	testData.PkB.Set(pkB)
	testData.ProofSkKnowledgeB.Set(proofSkKnowledgeB)
	testData.ProofCorrectDecryptionValid.Set(proofCorrectDecryptionValid)
	testData.ProofCorrectDecryptionInvalid.Set(proofCorrectDecryptionInvalid)
	testData.Result = int64(result)

	testDataJSON, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		panic(err)
	}
	println(string(testDataJSON))
}
