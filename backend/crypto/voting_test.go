package crypto

import (
	"crypto/rand"
	"errors"
	"io"
	mathrand "math/rand"
	"testing"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

func TestEncryptDecryptVote(t *testing.T) {
	tests := generateVotingTests()
	keyPair := generateKeyPair(t, rand.Reader)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			encryptedVote, _, err := tc.vote.Encrypt(rand.Reader, &keyPair.Pk)
			if err != nil {
				t.Fatal(err)
			}
			got, err := encryptedVote.Decrypt(&keyPair.Sk, tc.result)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.result {
				t.Fatalf("expected: %d, got: %d", tc.result, got)
			}
		})
	}
}

func TestTallying(t *testing.T) {
	tests := generateTallyingTests()
	keyPair := generateKeyPair(t, rand.Reader)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			encryptedResult := generateEncryptedResult(
				t,
				rand.Reader,
				&keyPair.Pk,
				tc.numVoters,
				tc.numYes)
			result, err := encryptedResult.Decrypt(&keyPair.Sk, tc.numVoters)
			if err != nil {
				t.Fatal(err)
			}
			if tc.numYes != result {
				t.Fatalf("expected: %d yes, got: %d", tc.numYes, result)
			}
		})
	}
}

func generateEncryptedResult(
	t *testing.T,
	r io.Reader,
	pk *arith.CurvePoint,
	numVoters int64,
	numYes int64) *EncryptedVote {

	if numYes > numVoters {
		t.Fatal(errors.New("numYes cannot be greater than numVoters"))
	}
	encryptedResult := NewEncryptedVote()
	perm := mathrand.Perm(int(numVoters))
	for _, i := range perm {
		var vote Vote
		if int64(i) < numYes {
			vote = Yes
		} else {
			vote = No
		}
		encryptedVote, _, err := vote.Encrypt(rand.Reader, pk)
		if err != nil {
			t.Fatal(err)
		}
		encryptedResult.Add(encryptedResult, encryptedVote)
	}
	return encryptedResult
}

type votingTests map[string]struct {
	vote   Vote
	result int64
}

func generateVotingTests() votingTests {
	tests := votingTests{
		"yes": {vote: Yes, result: 1},
		"no":  {vote: No, result: 0},
		"42":  {vote: 42, result: 42},
	}
	return tests
}

type tallyingTests map[string]struct {
	numVoters int64
	numYes    int64
}

func generateTallyingTests() tallyingTests {
	tests := tallyingTests{
		"1 voter, 0 yes":       {numVoters: 1, numYes: 0},
		"1 voter, 1 yes":       {numVoters: 1, numYes: 1},
		"2 voters, 0 yes":      {numVoters: 2, numYes: 0},
		"2 voters, 1 yes":      {numVoters: 2, numYes: 1},
		"2 voters, 2 yes":      {numVoters: 2, numYes: 2},
		"3 voters, 0 yes":      {numVoters: 3, numYes: 0},
		"3 voters, 1 yes":      {numVoters: 3, numYes: 1},
		"3 voters, 2 yes":      {numVoters: 3, numYes: 2},
		"3 voters, 3 yes":      {numVoters: 3, numYes: 3},
		"4 voters, 0 yes":      {numVoters: 4, numYes: 0},
		"4 voters, 4 yes":      {numVoters: 4, numYes: 4},
		"1000 voters, 313 yes": {numVoters: 1000, numYes: 313},
	}
	return tests
}
