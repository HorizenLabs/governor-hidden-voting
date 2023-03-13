package crypto

import (
	"errors"
	"io"
	"math/big"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

const numBytesProofVoteWellFormedness = 4*arith.NumBytesCurvePoint +
	2*arith.NumBytesScalar +
	arith.NumBytesChallenge

type ProofVoteWellFormedness struct {
	a0 arith.CurvePoint
	a1 arith.CurvePoint
	b0 arith.CurvePoint
	b1 arith.CurvePoint
	r0 arith.Scalar
	r1 arith.Scalar
	c0 arith.Challenge
}

func ProveVoteWellFormedness(
	reader io.Reader,
	encryptedVote *EncryptedVote,
	vote Vote,
	r *arith.Scalar,
	pk *arith.CurvePoint) (*ProofVoteWellFormedness, error) {
	// Generate cheating proof for the unchosen alternative
	var b *arith.CurvePoint
	switch vote {
	case Yes:
		b = &encryptedVote.b
	case No:
		b = new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-1)))
		b = new(arith.CurvePoint).Add(&encryptedVote.b, b)
	}
	cCheat, err := arith.RandomChallenge(reader)
	if err != nil {
		return nil, err
	}

	rCheat, rCheatG, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}

	cCheatNegA := new(arith.CurvePoint).ScalarMult(
		&encryptedVote.a, new(arith.Scalar).Neg(cCheat.Scalar()))
	aCheat := new(arith.CurvePoint).Add(rCheatG, cCheatNegA)

	cCheatNegB := new(arith.CurvePoint).ScalarMult(
		b, new(arith.Scalar).Neg(cCheat.Scalar()))
	rCheatPk := new(arith.CurvePoint).ScalarMult(pk, rCheat)
	bCheat := new(arith.CurvePoint).Add(rCheatPk, cCheatNegB)

	// Generate honest proof for chosen alternative
	rPrime, aHonest, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}
	bHonest := new(arith.CurvePoint).ScalarMult(pk, rPrime)

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesA, err := encryptedVote.a.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesB, err := encryptedVote.b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesACheat, err := aCheat.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesBCheat, err := bCheat.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesAHonest, err := aHonest.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesBHonest, err := bHonest.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var c *arith.Challenge
	switch vote {
	case Yes:
		c = arith.FiatShamirChallenge(
			bytesPk,
			bytesA,
			bytesB,
			bytesACheat,
			bytesBCheat,
			bytesAHonest,
			bytesBHonest)
	case No:
		c = arith.FiatShamirChallenge(
			bytesPk,
			bytesA,
			bytesB,
			bytesAHonest,
			bytesBHonest,
			bytesACheat,
			bytesBCheat)
	}

	cHonest := new(arith.Challenge).Sub(c, cCheat)

	cHonestR := new(arith.Scalar).Mul(cHonest.Scalar(), r)
	rHonest := new(arith.Scalar).Add(rPrime, cHonestR)

	proof := new(ProofVoteWellFormedness)
	switch vote {
	case Yes:
		proof.a0.Set(aCheat)
		proof.a1.Set(aHonest)
		proof.b0.Set(bCheat)
		proof.b1.Set(bHonest)
		proof.c0.Set(cCheat)
		proof.r0.Set(rCheat)
		proof.r1.Set(rHonest)
	case No:
		proof.a0.Set(aHonest)
		proof.a1.Set(aCheat)
		proof.b0.Set(bHonest)
		proof.b1.Set(bCheat)
		proof.c0.Set(cHonest)
		proof.r0.Set(rHonest)
		proof.r1.Set(rCheat)
	}
	return proof, nil
}

func VerifyVoteWellFormedness(
	proof *ProofVoteWellFormedness,
	vote *EncryptedVote,
	pk *arith.CurvePoint) error {

	m, err := proof.MarshalBinary()
	if err != nil {
		return err
	}
	proof = new(ProofVoteWellFormedness)
	err = proof.UnmarshalBinary(m)
	if err != nil {
		return err
	}

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA, err := vote.a.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB, err := vote.b.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA0, err := proof.a0.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB0, err := proof.b0.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA1, err := proof.a1.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB1, err := proof.b1.MarshalBinary()
	if err != nil {
		return err
	}

	c := arith.FiatShamirChallenge(
		bytesPk,
		bytesA,
		bytesB,
		bytesA0,
		bytesB0,
		bytesA1,
		bytesB1,
	)
	c1 := new(arith.Challenge).Sub(c, &proof.c0)

	r0G := new(arith.CurvePoint).ScalarBaseMult(&proof.r0)
	c0A := new(arith.CurvePoint).ScalarMult(&vote.a, proof.c0.Scalar())
	a0PlusC0A := new(arith.CurvePoint).Add(&proof.a0, c0A)
	if !r0G.Equal(a0PlusC0A) {
		return errors.New("vote well-formedness proof verification failed, first check")
	}

	r1G := new(arith.CurvePoint).ScalarBaseMult(&proof.r1)
	c1A := new(arith.CurvePoint).ScalarMult(&vote.a, c1.Scalar())
	a1PlusC1A := new(arith.CurvePoint).Add(&proof.a1, c1A)
	if !r1G.Equal(a1PlusC1A) {
		return errors.New("vote well-formedness proof verification failed, second check")
	}

	r0Pk := new(arith.CurvePoint).ScalarMult(pk, &proof.r0)
	c0B := new(arith.CurvePoint).ScalarMult(&vote.b, proof.c0.Scalar())
	b0PlusC0B := new(arith.CurvePoint).Add(&proof.b0, c0B)
	if !r0Pk.Equal(b0PlusC0B) {
		return errors.New("vote well-formedness proof verification failed, third check")
	}

	r1Pk := new(arith.CurvePoint).ScalarMult(pk, &proof.r1)
	gNeg := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-1)))
	bMinusG := new(arith.CurvePoint).Add(&vote.b, gNeg)
	c1BMinusG := new(arith.CurvePoint).ScalarMult(bMinusG, c1.Scalar())
	b1PlusC1BMinusG := new(arith.CurvePoint).Add(&proof.b1, c1BMinusG)

	if !r1Pk.Equal(b1PlusC1BMinusG) {
		return errors.New("vote well-formedness proof verification failed, fourth check")
	}
	return nil
}

func (proof *ProofVoteWellFormedness) MarshalBinary() ([]byte, error) {
	bytesA0, err := proof.a0.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesA1, err := proof.a1.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesB0, err := proof.b0.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesB1, err := proof.b1.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesR0, err := proof.r0.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesR1, err := proof.r1.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesC0, err := proof.c0.MarshalBinary()
	if err != nil {
		return nil, err
	}

	ret := make([]byte, 0, numBytesProofSkKnowledge)
	ret = append(ret, bytesA0...)
	ret = append(ret, bytesA1...)
	ret = append(ret, bytesB0...)
	ret = append(ret, bytesB1...)
	ret = append(ret, bytesR0...)
	ret = append(ret, bytesR1...)
	ret = append(ret, bytesC0...)
	return ret, nil
}

func (proof *ProofVoteWellFormedness) UnmarshalBinary(m []byte) error {
	if len(m) < numBytesProofVoteWellFormedness {
		return errors.New("ProofVoteWellFormedness: not enough data")
	}
	var err error

	if err = proof.a0.UnmarshalBinary(m); err != nil {
		return err
	}
	m = m[arith.NumBytesCurvePoint:]
	if err = proof.a1.UnmarshalBinary(m); err != nil {
		return err
	}
	m = m[arith.NumBytesCurvePoint:]
	if err = proof.b0.UnmarshalBinary(m); err != nil {
		return err
	}
	m = m[arith.NumBytesCurvePoint:]
	if err = proof.b1.UnmarshalBinary(m); err != nil {
		return err
	}
	m = m[arith.NumBytesCurvePoint:]
	if err = proof.r0.UnmarshalBinary(m); err != nil {
		return err
	}
	m = m[arith.NumBytesScalar:]
	if err = proof.r1.UnmarshalBinary(m); err != nil {
		return err
	}
	m = m[arith.NumBytesScalar:]
	if err = proof.c0.UnmarshalBinary(m); err != nil {
		return err
	}
	return nil
}
