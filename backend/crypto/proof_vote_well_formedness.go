package crypto

import (
	"encoding/json"
	"errors"
	"io"
	"math/big"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

// ProofVoteWellFormedness is a cryptographic proof that an encrypted
// vote is well-formed, i.e. it encodes a 0 (No) or a 1 (Yes).
type ProofVoteWellFormedness struct {
	A0 arith.CurvePoint
	A1 arith.CurvePoint
	B0 arith.CurvePoint
	B1 arith.CurvePoint
	R0 arith.Scalar
	R1 arith.Scalar
	C0 arith.Challenge
}

// ProveVoteWellFormedness generates a proof of well-formedness of an encrypted vote.
// In order to obtain a valid proof, the parameters should be obtained in the following
// way:
//
//	encryptedVote, r, err := vote.Encrypt(rand.Reader, pk)
func ProveVoteWellFormedness(
	reader io.Reader,
	encryptedVote *EncryptedVote,
	vote Vote,
	r *arith.Scalar,
	pk *arith.CurvePoint) (*ProofVoteWellFormedness, error) {
	// Implementation based on https://eprint.iacr.org/2016/765.pdf, section 4.5

	// Generate cheating proof for the unchosen alternative
	var b *arith.CurvePoint
	switch vote {
	case Yes:
		b = &encryptedVote.B
	case No:
		b = new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-1)))
		b = new(arith.CurvePoint).Add(&encryptedVote.B, b)
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
		&encryptedVote.A, new(arith.Scalar).Neg(cCheat.Scalar()))
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
	bytesA, err := encryptedVote.A.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesB, err := encryptedVote.B.MarshalBinary()
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
		proof.A0.Set(aCheat)
		proof.A1.Set(aHonest)
		proof.B0.Set(bCheat)
		proof.B1.Set(bHonest)
		proof.C0.Set(cCheat)
		proof.R0.Set(rCheat)
		proof.R1.Set(rHonest)
	case No:
		proof.A0.Set(aHonest)
		proof.A1.Set(aCheat)
		proof.B0.Set(bHonest)
		proof.B1.Set(bCheat)
		proof.C0.Set(cHonest)
		proof.R0.Set(rHonest)
		proof.R1.Set(rCheat)
	}
	return proof, nil
}

// VerifyVoteWellFormedness verifies a proof of well-formedness of an encrypted vote.
func VerifyVoteWellFormedness(
	proof *ProofVoteWellFormedness,
	vote *EncryptedVote,
	pk *arith.CurvePoint) error {
	// Implementation based on https://eprint.iacr.org/2016/765.pdf, section 4.5

	m, err := json.Marshal(proof)
	if err != nil {
		return err
	}
	proof = new(ProofVoteWellFormedness)
	err = json.Unmarshal(m, proof)
	if err != nil {
		return err
	}

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA, err := vote.A.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB, err := vote.B.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA0, err := proof.A0.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB0, err := proof.B0.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA1, err := proof.A1.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB1, err := proof.B1.MarshalBinary()
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
	c1 := new(arith.Challenge).Sub(c, &proof.C0)

	r0G := new(arith.CurvePoint).ScalarBaseMult(&proof.R0)
	c0A := new(arith.CurvePoint).ScalarMult(&vote.A, proof.C0.Scalar())
	a0PlusC0A := new(arith.CurvePoint).Add(&proof.A0, c0A)
	if !r0G.Equal(a0PlusC0A) {
		return errors.New("vote well-formedness proof verification failed, first check")
	}

	r1G := new(arith.CurvePoint).ScalarBaseMult(&proof.R1)
	c1A := new(arith.CurvePoint).ScalarMult(&vote.A, c1.Scalar())
	a1PlusC1A := new(arith.CurvePoint).Add(&proof.A1, c1A)
	if !r1G.Equal(a1PlusC1A) {
		return errors.New("vote well-formedness proof verification failed, second check")
	}

	r0Pk := new(arith.CurvePoint).ScalarMult(pk, &proof.R0)
	c0B := new(arith.CurvePoint).ScalarMult(&vote.B, proof.C0.Scalar())
	b0PlusC0B := new(arith.CurvePoint).Add(&proof.B0, c0B)
	if !r0Pk.Equal(b0PlusC0B) {
		return errors.New("vote well-formedness proof verification failed, third check")
	}

	r1Pk := new(arith.CurvePoint).ScalarMult(pk, &proof.R1)
	gNeg := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-1)))
	bMinusG := new(arith.CurvePoint).Add(&vote.B, gNeg)
	c1BMinusG := new(arith.CurvePoint).ScalarMult(bMinusG, c1.Scalar())
	b1PlusC1BMinusG := new(arith.CurvePoint).Add(&proof.B1, c1BMinusG)

	if !r1Pk.Equal(b1PlusC1BMinusG) {
		return errors.New("vote well-formedness proof verification failed, fourth check")
	}
	return nil
}
