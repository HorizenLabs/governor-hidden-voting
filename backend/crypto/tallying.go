package crypto

import (
	"math/big"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

type Tally struct {
	NumYes int64
	NumNo  int64
}

type EncryptedTally struct {
	Votes EncryptedVote
	Count int64
}

// NewEncryptedTally allocates and returns an EncryptedTally encoding zero votes.
func NewEncryptedTally() *EncryptedTally {
	tally := new(EncryptedTally)
	zero := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(0)))
	tally.Votes.A.Set(zero)
	tally.Votes.B.Set(zero)
	tally.Count = 0
	return tally
}

func (tally *Tally) NumVoters() int64 {
	return tally.NumYes + tally.NumNo
}

func (tally *EncryptedTally) Add(vote *EncryptedVote) *EncryptedTally {
	tally.Votes.Add(&tally.Votes, vote)
	tally.Count++
	return tally
}

func (tally *EncryptedTally) Decrypt(sk *arith.Scalar) (*Tally, error) {
	numYes, err := tally.Votes.Decrypt(sk, tally.Count)
	if err != nil {
		return nil, err
	}
	return &Tally{NumYes: numYes, NumNo: tally.Count - numYes}, nil
}
