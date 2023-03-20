package crypto

import (
	"crypto/rand"
	"errors"
	"io"
	mathrand "math/rand"
	"testing"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

func TestTallying(t *testing.T) {
	tests := generateTallyingTests()
	keyPair := generateKeyPair(t, rand.Reader)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			encryptedTally := generateEncryptedTally(
				t,
				rand.Reader,
				&keyPair.Pk,
				tc.numVoters,
				tc.numYes)
			tally, err := encryptedTally.Decrypt(&keyPair.Sk)
			if err != nil {
				t.Fatal(err)
			}
			if tc.numYes != tally.NumYes {
				t.Fatalf("expected: %d yes, got: %d", tc.numYes, tally.NumYes)
			}
			numNo := tc.numVoters - tc.numYes
			if numNo != tally.NumNo {
				t.Fatalf("expected %d no, got: %d", numNo, tally.NumNo)
			}
		})
	}
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

func generateEncryptedTally(
	t *testing.T,
	r io.Reader,
	pk *arith.CurvePoint,
	numVoters int64,
	numYes int64) *EncryptedTally {

	if numYes > numVoters {
		t.Fatal(errors.New("numYes cannot be greater than numVoters"))
	}
	tally := NewEncryptedTally()
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
		tally.Add(encryptedVote)
	}
	return tally
}
