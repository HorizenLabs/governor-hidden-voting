package crypto

import (
	"fmt"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

type SmartContractMock struct {
	pk             *arith.CurvePoint
	encryptedTally *EncryptedVote
	result         uint64
	status         Status
}

type Status int

const (
	Init Status = iota
	Declared
	Voting
	Tallying
	Fini
)

func NewSmartContractMock() *SmartContractMock {
	return &SmartContractMock{
		encryptedTally: NewEncryptedVote(),
		status:         Init,
	}
}

func (sc *SmartContractMock) DeclarePk(pk *arith.CurvePoint, proof *ProofSkKnowledge) error {
	if sc.status != Init {
		return fmt.Errorf("wrong status")
	}
	err := VerifySkKnowledge(proof, pk)
	if err == nil {
		sc.pk = pk
		sc.status = Declared
	}
	return err
}

func (sc *SmartContractMock) StartVotingPhase() error {
	if sc.status != Declared {
		return fmt.Errorf("wrong status")
	}
	sc.status = Voting
	return nil
}

func (sc *SmartContractMock) CastVote(proof *ProofVoteWellFormedness, vote *EncryptedVote) error {
	if sc.status != Voting {
		return fmt.Errorf("wrong status")
	}
	err := VerifyVoteWellFormedness(proof, vote, sc.pk)
	if err == nil {
		sc.encryptedTally.Add(sc.encryptedTally, vote)
	}
	return err
}

func (sc *SmartContractMock) StopVotingPhase() error {
	if sc.status != Voting {
		return fmt.Errorf("wrong status")
	}
	sc.status = Tallying
	return nil
}

func (sc *SmartContractMock) Tally(proof *ProofCorrectDecryption, decryptedTally uint64) error {
	if sc.status != Tallying {
		return fmt.Errorf("wrong status")
	}
	err := VerifyCorrectDecryption(proof, sc.encryptedTally, Vote(decryptedTally), sc.pk)
	if err == nil {
		sc.result = decryptedTally
		sc.status = Fini
	}
	return err
}

func (sc *SmartContractMock) GetPk() (*arith.CurvePoint, error) {
	if sc.status == Init {
		return nil, fmt.Errorf("wrong status")
	}
	return sc.pk, nil
}

func (sc *SmartContractMock) GetResult() (uint64, error) {
	if sc.status != Fini {
		return 0, fmt.Errorf("wrong status")
	}
	return sc.result, nil
}

func (sc *SmartContractMock) GetEncryptedTally() (*EncryptedVote, error) {
	return sc.encryptedTally, nil
}
