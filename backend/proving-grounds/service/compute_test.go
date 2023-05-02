package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"testing"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
	"github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/capabilities"
	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type TestData struct {
	keyPairA                *crypto.KeyPair
	keyPairB                *crypto.KeyPair
	zeroEncryptedWithPkA    *crypto.EncryptedVote
	oneEncryptedWithPkA     *crypto.EncryptedVote
	zeroEncryptedWithPkB    *crypto.EncryptedVote
	oneEncryptedWithPkB     *crypto.EncryptedVote
	nonExistingCapabilityId string
}

func generateTestData() (*TestData, error) {
	rng := rand.New(rand.NewSource(42))

	testData := new(TestData)

	keyPairA, err := crypto.NewKeyPair(rng)
	if err != nil {
		return nil, err
	}
	testData.keyPairA = keyPairA

	keyPairB, err := crypto.NewKeyPair(rng)
	if err != nil {
		return nil, err
	}
	testData.keyPairB = keyPairB

	zeroEncryptedWithPkA, _, err := crypto.EncryptVoteWithProof(rng, 0, &keyPairA.Pk)
	if err != nil {
		return nil, err
	}
	testData.zeroEncryptedWithPkA = zeroEncryptedWithPkA

	zeroEncryptedWithPkB, _, err := crypto.EncryptVoteWithProof(rng, 0, &keyPairB.Pk)
	if err != nil {
		return nil, err
	}
	testData.zeroEncryptedWithPkB = zeroEncryptedWithPkB

	oneEncryptedWithPkA, _, err := crypto.EncryptVoteWithProof(rng, 1, &keyPairA.Pk)
	if err != nil {
		return nil, err
	}
	testData.oneEncryptedWithPkA = oneEncryptedWithPkA

	oneEncryptedWithPkB, _, err := crypto.EncryptVoteWithProof(rng, 1, &keyPairB.Pk)
	if err != nil {
		return nil, err
	}
	testData.oneEncryptedWithPkB = oneEncryptedWithPkB

	nonExistingCapabilityId, err := uuid.NewRandomFromReader(rng)
	if err != nil {
		return nil, err
	}
	testData.nonExistingCapabilityId = nonExistingCapabilityId.String()

	return testData, nil
}

type computeTests map[string]struct {
	capabilityId  string
	arguments     [][]byte
	shouldSucceed bool
	numOutputs    int
	checkResult   func(*cs.Result) error
}

func TestCompute(t *testing.T) {
	testData, err := generateTestData()
	if err != nil {
		t.Fatal(err)
	}

	computeJson := func(in any) []byte {
		ret, err := json.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		return ret
	}

	tests := computeTests{
		"NewKeyPairWithProof": {
			capabilityId:  capabilities.NEW_KEY_PAIR_WITH_PROOF_ID,
			arguments:     [][]byte{},
			shouldSucceed: true,
			numOutputs:    2,
			checkResult:   func(result *cs.Result) error { return checkNewKeyPairWithProofOutput(result) },
		},
		"EncryptVoteWithProof/0": {
			capabilityId:  capabilities.ENCRYPT_VOTE_WITH_PROOF_ID,
			arguments:     [][]byte{computeJson(0), computeJson(testData.keyPairA.Pk)},
			shouldSucceed: true,
			numOutputs:    2,
			checkResult:   func(result *cs.Result) error { return checkEncryptVoteWithProofOutput(result, &testData.keyPairA.Pk) },
		},
		"EncryptVoteWithProof/1": {
			capabilityId:  capabilities.ENCRYPT_VOTE_WITH_PROOF_ID,
			arguments:     [][]byte{computeJson(1), computeJson(testData.keyPairA.Pk)},
			shouldSucceed: true,
			numOutputs:    2,
			checkResult:   func(result *cs.Result) error { return checkEncryptVoteWithProofOutput(result, &testData.keyPairA.Pk) },
		},
		"DecryptTallyWithProof": {
			capabilityId:  capabilities.DECRYPT_TALLY_WITH_PROOF_ID,
			arguments:     [][]byte{computeJson(testData.oneEncryptedWithPkA), computeJson(1), computeJson(testData.keyPairA)},
			shouldSucceed: true,
			numOutputs:    2,
			checkResult: func(result *cs.Result) error {
				return checkDecryptTallyWithProofOutput(result, 1, testData.oneEncryptedWithPkA, &testData.keyPairA.Pk)
			},
		},
		"NonExistingCapability": {
			capabilityId:  testData.nonExistingCapabilityId,
			arguments:     [][]byte{},
			shouldSucceed: false,
			numOutputs:    0,
			checkResult:   func(result *cs.Result) error { return nil },
		},
		"EncryptVoteWithProof/MalformedVote": {
			capabilityId:  capabilities.ENCRYPT_VOTE_WITH_PROOF_ID,
			arguments:     [][]byte{computeJson(2), computeJson(testData.keyPairA.Pk)},
			shouldSucceed: false,
			numOutputs:    0,
			checkResult:   func(result *cs.Result) error { return nil },
		},
		"DecryptTallyWithProof/WrongUpperBound": {
			capabilityId:  capabilities.DECRYPT_TALLY_WITH_PROOF_ID,
			arguments:     [][]byte{computeJson(testData.oneEncryptedWithPkA), computeJson(0), computeJson(testData.keyPairA)},
			shouldSucceed: false,
			numOutputs:    0,
			checkResult:   func(result *cs.Result) error { return nil },
		},
		"DecryptTallyWithProof/WrongPk": {
			capabilityId:  capabilities.DECRYPT_TALLY_WITH_PROOF_ID,
			arguments:     [][]byte{computeJson(testData.oneEncryptedWithPkA), computeJson(1), computeJson(testData.keyPairB)},
			shouldSucceed: false,
			numOutputs:    0,
			checkResult:   func(result *cs.Result) error { return nil },
		},
	}

	ctx := context.Background()
	client, closer := server(ctx)
	defer closer()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := compute(client, ctx, tc.capabilityId, tc.arguments)
			if err != nil {
				t.Fatal(err)
			}

			if tc.shouldSucceed {
				if output.GetError() != "" {
					t.Fatalf("expected empty Error field, instead got \"%v\"", output.GetError())
				}
				if !output.Ok {
					t.Fatalf("Ok field should be true")
				}
				if len(output.GetResult()) != tc.numOutputs {
					t.Fatalf("expected %d outputs, got %d", tc.numOutputs, len(output.GetResult()))
				}
				if err := tc.checkResult(output); err != nil {
					t.Fatal(err)
				}
			} else {
				if output.Ok {
					t.Fatalf("Ok field should be false")
				}
				if output.GetError() == "" {
					t.Fatalf("Error field should not be empty")
				}
				if len(output.GetResult()) != 0 {
					t.Fatalf("Result field should be empty")
				}
			}
		})
	}
}

func compute(client cs.CryptographicServiceClient, ctx context.Context, capabilityId string, arguments [][]byte) (*cs.Result, error) {
	requestUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("fail to compute uuid for request: %v", err)
	}
	in := &cs.Execute{
		RequestId:    requestUUID.String(),
		CapabilityId: capabilityId,
		Arguments:    arguments,
	}

	output, err := client.Compute(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("client.Compute failed: %v", err)
	}

	return output, nil
}

func checkNewKeyPairWithProofOutput(output *cs.Result) error {
	keyPair := new(crypto.KeyPair)
	if err := json.Unmarshal(output.GetResult()[0], keyPair); err != nil {
		return err
	}

	proofSkKnowledge := new(crypto.ProofSkKnowledge)
	if err := json.Unmarshal(output.GetResult()[1], proofSkKnowledge); err != nil {
		return err
	}

	if err := crypto.VerifySkKnowledge(proofSkKnowledge, &keyPair.Pk); err != nil {
		return err
	}

	return nil
}

func checkEncryptVoteWithProofOutput(output *cs.Result, pk *arith.CurvePoint) error {
	encryptedVote := new(crypto.EncryptedVote)
	if err := json.Unmarshal(output.GetResult()[0], encryptedVote); err != nil {
		return err
	}

	proofVoteWellFormedness := new(crypto.ProofVoteWellFormedness)
	if err := json.Unmarshal(output.GetResult()[1], proofVoteWellFormedness); err != nil {
		return err
	}

	if err := crypto.VerifyVoteWellFormedness(proofVoteWellFormedness, encryptedVote, pk); err != nil {
		return err
	}

	return nil
}

func checkDecryptTallyWithProofOutput(output *cs.Result, vote int64, encryptedVote *crypto.EncryptedVote, pk *arith.CurvePoint) error {
	decryptedVote := new(int64)
	if err := json.Unmarshal(output.GetResult()[0], decryptedVote); err != nil {
		return err
	}
	if *decryptedVote != vote {
		return fmt.Errorf("wrong decrypted vote: expected %d, got %d", vote, *decryptedVote)
	}

	proofCorrectDecryption := new(crypto.ProofCorrectDecryption)
	if err := json.Unmarshal(output.GetResult()[1], proofCorrectDecryption); err != nil {
		return err
	}

	if err := crypto.VerifyCorrectDecryption(proofCorrectDecryption, encryptedVote, crypto.Vote(*decryptedVote), pk); err != nil {
		return err
	}

	return nil
}

func server(ctx context.Context) (cs.CryptographicServiceClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	grpcServer := grpc.NewServer()
	cs.RegisterCryptographicServiceServer(grpcServer, newServer())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		grpcServer.Stop()
	}

	client := cs.NewCryptographicServiceClient(conn)

	return client, closer
}
