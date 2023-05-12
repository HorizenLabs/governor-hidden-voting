package main

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
	"golang.org/x/net/nettest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestElection(t *testing.T) {
	smartContract := crypto.NewSmartContractMock()
	backend, closer := newEVotingBackend(t)
	defer closer()

	// Election authority generates a keypair and declares election pk
	keyPair, proofSkKnowledge := backend.RequestNewKeyPairWithProof()
	if err := smartContract.DeclarePk(&keyPair.Pk, proofSkKnowledge); err != nil {
		t.Fatal(err)
	}

	// Election authority starts voting phase
	if err := smartContract.StartVotingPhase(); err != nil {
		t.Fatal(err)
	}

	// Voters cast their votes
	votes := []uint64{1, 0, 0, 1, 0}
	for _, vote := range votes {
		pk, err := smartContract.GetPk()
		if err != nil {
			t.Fatal(err)
		}
		encryptedVote, proofVoteWellFormedness := backend.RequestEncryptVoteWithProof(vote, pk)
		if err := smartContract.CastVote(proofVoteWellFormedness, encryptedVote); err != nil {
			t.Fatal(err)
		}
	}

	// Election authority stops voting phase
	if err := smartContract.StopVotingPhase(); err != nil {
		t.Fatal(err)
	}

	// Election authority performs tallying
	numVoters := len(votes)
	encryptedTally, err := smartContract.GetEncryptedTally()
	if err != nil {
		t.Fatal(err)
	}
	decryptedTally, proofCorrectDecryption := backend.RequestDecryptTallyWithProof(
		encryptedTally,
		uint64(numVoters),
		keyPair,
	)
	if err := smartContract.Tally(proofCorrectDecryption, decryptedTally); err != nil {
		t.Fatal(err)
	}

	// check that result is correct
	correctResult := uint64(0)
	for _, vote := range votes {
		correctResult += vote
	}
	result, err := smartContract.GetResult()
	if err != nil {
		t.Fatal(err)
	}
	if result != correctResult {
		t.Fatalf("result: expected %d, got %d", correctResult, result)
	}
}

type eVotingBackend struct {
	client cs.CryptographicServiceClient
	t      *testing.T
}

func newEVotingBackend(t *testing.T) (*eVotingBackend, func()) {
	client, closer := eVotingServer(context.Background())
	backend := &eVotingBackend{
		client: client,
		t:      t,
	}
	return backend, closer
}

func (e *eVotingBackend) RequestNewKeyPairWithProof() (*crypto.KeyPair, *crypto.ProofSkKnowledge) {
	newKeyPairRequest := &cs.Execute{
		CapabilityId: "924e949e-0f98-42c6-a4ff-bb2b1c8693e0",
	}
	newKeyPairWithProofResponse, err := e.client.Compute(context.Background(), newKeyPairRequest)
	if err != nil {
		e.t.Fatal(err)
	}
	if !newKeyPairWithProofResponse.GetOk() {
		e.t.Fatalf(newKeyPairWithProofResponse.GetError())
	}
	keyPairWithProofJSON := newKeyPairWithProofResponse.GetResult()
	keyPair := new(crypto.KeyPair)
	if err := json.Unmarshal(keyPairWithProofJSON[0], keyPair); err != nil {
		e.t.Fatal(err)
	}
	proofSkKnowledge := new(crypto.ProofSkKnowledge)
	if err := json.Unmarshal(keyPairWithProofJSON[1], proofSkKnowledge); err != nil {
		e.t.Fatal(err)
	}
	return keyPair, proofSkKnowledge
}

func (e *eVotingBackend) RequestEncryptVoteWithProof(
	vote uint64,
	pk *arith.CurvePoint,
) (*crypto.EncryptedVote, *crypto.ProofVoteWellFormedness) {
	voteJSON, err := json.Marshal(vote)
	if err != nil {
		e.t.Fatal(err)
	}
	pkJSON, err := json.Marshal(pk)
	if err != nil {
		e.t.Fatal(err)
	}

	encryptVoteRequest := &cs.Execute{
		CapabilityId: "772e6272-8c49-463f-b2d1-92088ae06da1",
		Arguments:    [][]byte{voteJSON, pkJSON},
	}
	encryptVoteResponse, err := e.client.Compute(context.Background(), encryptVoteRequest)
	if err != nil {
		e.t.Fatal(err)
	}
	if !encryptVoteResponse.GetOk() {
		e.t.Fatalf(encryptVoteResponse.GetError())
	}
	encryptedVoteWithProofJSON := encryptVoteResponse.GetResult()
	encryptedVote := new(crypto.EncryptedVote)
	if err := json.Unmarshal(encryptedVoteWithProofJSON[0], encryptedVote); err != nil {
		e.t.Fatal(err)
	}
	proofVoteWellFormedness := new(crypto.ProofVoteWellFormedness)
	if err := json.Unmarshal(encryptedVoteWithProofJSON[1], proofVoteWellFormedness); err != nil {
		e.t.Fatal(err)
	}
	return encryptedVote, proofVoteWellFormedness
}

func (e *eVotingBackend) RequestDecryptTallyWithProof(
	encryptedTally *crypto.EncryptedVote,
	numVoters uint64,
	keyPair *crypto.KeyPair,
) (uint64, *crypto.ProofCorrectDecryption) {
	encryptedTallyJSON, err := json.Marshal(encryptedTally)
	if err != nil {
		e.t.Fatal(err)
	}
	numVotersJSON, err := json.Marshal(numVoters)
	if err != nil {
		e.t.Fatal(err)
	}
	keyPairJSON, err := json.Marshal(keyPair)
	if err != nil {
		e.t.Fatal(err)
	}
	decryptTallyRequest := &cs.Execute{
		CapabilityId: "c238d864-ae22-4db5-b3d1-83c41cb8b4dd",
		Arguments:    [][]byte{encryptedTallyJSON, numVotersJSON, keyPairJSON},
	}
	decryptTallyWithProofResponse, err := e.client.Compute(context.Background(), decryptTallyRequest)
	if err != nil {
		e.t.Fatal(err)
	}
	if !decryptTallyWithProofResponse.GetOk() {
		e.t.Fatalf(decryptTallyWithProofResponse.GetError())
	}
	decryptTallyWithProofJSON := decryptTallyWithProofResponse.GetResult()
	result := new(uint64)
	if err := json.Unmarshal(decryptTallyWithProofJSON[0], result); err != nil {
		e.t.Fatal(err)
	}
	proofCorrectDecryption := new(crypto.ProofCorrectDecryption)
	if err := json.Unmarshal(decryptTallyWithProofJSON[1], proofCorrectDecryption); err != nil {
		e.t.Fatal(err)
	}
	return *result, proofCorrectDecryption
}

func eVotingServer(ctx context.Context) (cs.CryptographicServiceClient, func()) {
	lis, err := nettest.NewLocalListener("tcp")
	if err != nil {
		log.Print(err)
	}

	baseServer := grpc.NewServer()
	service := newEvotingService()
	cs.RegisterCryptographicServiceServer(baseServer, service)

	go func() {
		if err := service.Serve(lis); err != nil {
			log.Print(err)
		}
	}()

	var dialOpts []grpc.DialOption
	dialOpts = append(
		dialOpts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	conn, err := grpc.DialContext(ctx, lis.Addr().String(), dialOpts...)
	if err != nil {
		log.Print(err)
	}

	closer := func() {
		if err := lis.Close(); err != nil {
			log.Print(err)
		}
		baseServer.Stop()
	}

	client := cs.NewCryptographicServiceClient(conn)

	return client, closer
}
