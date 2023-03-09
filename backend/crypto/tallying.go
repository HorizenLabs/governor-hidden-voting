package crypto

import (
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

func NewEncryptedTally() *EncryptedTally {
	tally := new(EncryptedTally)
	tally.Count = 0
	return tally
}

func (tally *Tally) NumVoters() int64 {
	return tally.NumYes + tally.NumNo
}

func (tally *EncryptedTally) Add(vote *EncryptedVote) *EncryptedTally {
	if tally.Count == 0 {
		tally.Votes.Set(vote)
	} else {
		tally.Votes.Add(&tally.Votes, vote)
	}
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
