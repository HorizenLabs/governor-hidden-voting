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
	R0 arith.Scalar
	R1 arith.Scalar
	C0 arith.Challenge
	C1 arith.Challenge
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
	default:
		return nil, errors.New("proof of vote well formedness can only be generated for yes/no vote")
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
		proof.R0.Set(rCheat)
		proof.R1.Set(rHonest)
		proof.C0.Set(cCheat)
		proof.C1.Set(cHonest)
	case No:
		proof.R0.Set(rHonest)
		proof.R1.Set(rCheat)
		proof.C0.Set(cHonest)
		proof.C1.Set(cCheat)
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

	r0G := new(arith.CurvePoint).ScalarBaseMult(&proof.R0)
	c0A := new(arith.CurvePoint).ScalarMult(&vote.A, proof.C0.Scalar())
	a0 := new(arith.CurvePoint).Add(r0G, new(arith.CurvePoint).Neg(c0A))

	r1G := new(arith.CurvePoint).ScalarBaseMult(&proof.R1)
	c1A := new(arith.CurvePoint).ScalarMult(&vote.A, proof.C1.Scalar())
	a1 := new(arith.CurvePoint).Add(r1G, new(arith.CurvePoint).Neg(c1A))

	r0Pk := new(arith.CurvePoint).ScalarMult(pk, &proof.R0)
	c0B := new(arith.CurvePoint).ScalarMult(&vote.B, proof.C0.Scalar())
	b0 := new(arith.CurvePoint).Add(r0Pk, new(arith.CurvePoint).Neg(c0B))

	r1Pk := new(arith.CurvePoint).ScalarMult(pk, &proof.R1)
	gNeg := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-1)))
	bMinusG := new(arith.CurvePoint).Add(&vote.B, gNeg)
	c1BMinusG := new(arith.CurvePoint).ScalarMult(bMinusG, proof.C1.Scalar())
	b1 := new(arith.CurvePoint).Add(r1Pk, new(arith.CurvePoint).Neg(c1BMinusG))

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
	bytesA0, err := a0.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB0, err := b0.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA1, err := a1.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB1, err := b1.MarshalBinary()
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

	if !c.Equal(new(arith.Challenge).Add(&proof.C0, &proof.C1)) {
		return errors.New("vote well-formedness proof verification failed")
	}
	return nil
}
