package crypto

import (
	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

type Tally struct {
	numYes int64
	numNo  int64
}

type EncryptedTally struct {
	votes EncryptedVote
	count int64
}

func NewEncryptedTally() *EncryptedTally {
	tally := new(EncryptedTally)
	tally.count = 0
	return tally
}

func (tally *Tally) NumYes() int64 {
	return tally.numYes
}

func (tally *Tally) NumNo() int64 {
	return tally.numNo
}

func (tally *Tally) NumVoters() int64 {
	return tally.numYes + tally.numNo
}

func (tally *EncryptedTally) Add(vote *EncryptedVote) *EncryptedTally {
	if tally.count == 0 {
		tally.votes.Set(vote)
	} else {
		tally.votes.Add(&tally.votes, vote)
	}
	tally.count++
	return tally
}

func (tally *EncryptedTally) Decrypt(sk *arith.Scalar) (*Tally, error) {
	numYes, err := tally.votes.Decrypt(sk, tally.count)
	if err != nil {
		return nil, err
	}
	return &Tally{numYes: *numYes, numNo: tally.count - *numYes}, nil
}
