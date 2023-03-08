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
	a0 *arith.CurvePoint
	a1 *arith.CurvePoint
	b0 *arith.CurvePoint
	b1 *arith.CurvePoint
	r0 *arith.Scalar
	r1 *arith.Scalar
	c0 *arith.Challenge
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
		b = encryptedVote.b
	case No:
		b = new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-1)))
		b = new(arith.CurvePoint).Add(encryptedVote.b, b)
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
		encryptedVote.a, new(arith.Scalar).Neg(cCheat.Scalar()))
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

	var c *arith.Challenge
	switch vote {
	case Yes:
		c = arith.FiatShamirChallenge(
			pk.Marshal(),
			encryptedVote.a.Marshal(),
			encryptedVote.b.Marshal(),
			aCheat.Marshal(),
			bCheat.Marshal(),
			aHonest.Marshal(),
			bHonest.Marshal())
	case No:
		c = arith.FiatShamirChallenge(
			pk.Marshal(),
			encryptedVote.a.Marshal(),
			encryptedVote.b.Marshal(),
			aHonest.Marshal(),
			bHonest.Marshal(),
			aCheat.Marshal(),
			bCheat.Marshal())
	}

	cHonest := new(arith.Challenge).Sub(c, cCheat)

	cHonestR := new(arith.Scalar).Mul(cHonest.Scalar(), r)
	rHonest := new(arith.Scalar).Add(rPrime, cHonestR)

	var proof ProofVoteWellFormedness
	switch vote {
	case Yes:
		proof = ProofVoteWellFormedness{
			a0: aCheat,
			a1: aHonest,
			b0: bCheat,
			b1: bHonest,
			c0: cCheat,
			r0: rCheat,
			r1: rHonest,
		}
	case No:
		proof = ProofVoteWellFormedness{
			a0: aHonest,
			a1: aCheat,
			b0: bHonest,
			b1: bCheat,
			c0: cHonest,
			r0: rHonest,
			r1: rCheat,
		}
	}
	return &proof, nil
}

func VerifyVoteWellFormedness(
	proof *ProofVoteWellFormedness,
	vote *EncryptedVote,
	pk *arith.CurvePoint) error {

	m := proof.Marshal()
	proof = new(ProofVoteWellFormedness)
	proof.Unmarshal(m)

	c := arith.FiatShamirChallenge(
		pk.Marshal(),
		vote.a.Marshal(),
		vote.b.Marshal(),
		proof.a0.Marshal(),
		proof.b0.Marshal(),
		proof.a1.Marshal(),
		proof.b1.Marshal(),
	)
	c1 := new(arith.Challenge).Sub(c, proof.c0)

	r0G := new(arith.CurvePoint).ScalarBaseMult(proof.r0)
	c0A := new(arith.CurvePoint).ScalarMult(vote.a, proof.c0.Scalar())
	a0PlusC0A := new(arith.CurvePoint).Add(proof.a0, c0A)
	if !r0G.Equal(a0PlusC0A) {
		return errors.New("vote well-formedness proof verification failed, first check")
	}

	r1G := new(arith.CurvePoint).ScalarBaseMult(proof.r1)
	c1A := new(arith.CurvePoint).ScalarMult(vote.a, c1.Scalar())
	a1PlusC1A := new(arith.CurvePoint).Add(proof.a1, c1A)
	if !r1G.Equal(a1PlusC1A) {
		return errors.New("vote well-formedness proof verification failed, second check")
	}

	r0Pk := new(arith.CurvePoint).ScalarMult(pk, proof.r0)
	c0B := new(arith.CurvePoint).ScalarMult(vote.b, proof.c0.Scalar())
	b0PlusC0B := new(arith.CurvePoint).Add(proof.b0, c0B)
	if !r0Pk.Equal(b0PlusC0B) {
		return errors.New("vote well-formedness proof verification failed, third check")
	}

	r1Pk := new(arith.CurvePoint).ScalarMult(pk, proof.r1)
	gNeg := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-1)))
	bMinusG := new(arith.CurvePoint).Add(vote.b, gNeg)
	c1BMinusG := new(arith.CurvePoint).ScalarMult(bMinusG, c1.Scalar())
	b1PlusC1BMinusG := new(arith.CurvePoint).Add(proof.b1, c1BMinusG)

	if !r1Pk.Equal(b1PlusC1BMinusG) {
		return errors.New("vote well-formedness proof verification failed, fourth check")
	}
	return nil
}

func (proof *ProofVoteWellFormedness) Marshal() []byte {
	ret := make([]byte, 0, numBytesProofVoteWellFormedness)
	ret = append(ret, proof.a0.Marshal()...)
	ret = append(ret, proof.a1.Marshal()...)
	ret = append(ret, proof.b0.Marshal()...)
	ret = append(ret, proof.b1.Marshal()...)
	ret = append(ret, proof.r0.Marshal()...)
	ret = append(ret, proof.r1.Marshal()...)
	ret = append(ret, proof.c0.Marshal()...)
	return ret
}

func (proof *ProofVoteWellFormedness) Unmarshal(m []byte) ([]byte, error) {
	if len(m) < numBytesProofVoteWellFormedness {
		return nil, errors.New("ProofVoteWellFormedness: not enough data")
	}
	var err error

	if proof.a0 == nil {
		proof.a0 = &arith.CurvePoint{}
	}
	if m, err = proof.a0.Unmarshal(m); err != nil {
		return nil, err
	}

	if proof.a1 == nil {
		proof.a1 = &arith.CurvePoint{}
	}
	if m, err = proof.a1.Unmarshal(m); err != nil {
		return nil, err
	}

	if proof.b0 == nil {
		proof.b0 = &arith.CurvePoint{}
	}
	if m, err = proof.b0.Unmarshal(m); err != nil {
		return nil, err
	}

	if proof.b1 == nil {
		proof.b1 = &arith.CurvePoint{}
	}
	if m, err = proof.b1.Unmarshal(m); err != nil {
		return nil, err
	}

	if proof.r0 == nil {
		proof.r0 = &arith.Scalar{}
	}
	if m, err = proof.r0.Unmarshal(m); err != nil {
		return nil, err
	}

	if proof.r1 == nil {
		proof.r1 = &arith.Scalar{}
	}
	if m, err = proof.r1.Unmarshal(m); err != nil {
		return nil, err
	}

	if proof.c0 == nil {
		proof.c0 = &arith.Challenge{}
	}
	if m, err = proof.c0.Unmarshal(m); err != nil {
		return nil, err
	}

	return m, nil
}
