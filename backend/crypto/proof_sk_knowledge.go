package crypto

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

// ProofSkKnowledge is a cryptographic proof of knowledge of the secret key of an ElGamal KeyPair
type ProofSkKnowledge struct {
	D arith.Scalar
	C arith.Challenge
}

// ProveSkKnowledge generates a proof of knowledge of the secret key of an ElGamal KeyPair
func ProveSkKnowledge(reader io.Reader, keyPair *KeyPair) (*ProofSkKnowledge, error) {
	// Implementation based on https://eprint.iacr.org/2016/765.pdf, section 4.3
	r, b, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}
	bytesPk, err := keyPair.Pk.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesB, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	c := arith.FiatShamirChallenge(bytesPk, bytesB)
	d := new(arith.Scalar).Mul(c.Scalar(), &keyPair.Sk)
	d = new(arith.Scalar).Add(r, d)
	proof := new(ProofSkKnowledge)
	proof.D.Set(d)
	proof.C.Set(c)
	return proof, nil
}

// VerifySkKnowledge verifies a proof of knowledge of the secret key of an ElGamal KeyPair
func VerifySkKnowledge(proof *ProofSkKnowledge, pk *arith.CurvePoint) error {
	// Implementation based on https://eprint.iacr.org/2016/765.pdf, section 4.3
	m, err := json.Marshal(proof)
	if err != nil {
		return err
	}
	proof = new(ProofSkKnowledge)
	err = json.Unmarshal(m, proof)
	if err != nil {
		return err
	}

	cPk := new(arith.CurvePoint).ScalarMult(pk, proof.C.Scalar())
	b := new(arith.CurvePoint).ScalarBaseMult(&proof.D)
	b.Add(b, new(arith.CurvePoint).Neg(cPk))

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB, err := b.MarshalBinary()
	if err != nil {
		return err
	}
	c := arith.FiatShamirChallenge(bytesPk, bytesB)
	if !c.Equal(&proof.C) {
		return errors.New("sk knowledge proof verification failed, first check")
	}

	bPlusCPk := new(arith.CurvePoint).Add(b, cPk)
	phi := new(arith.CurvePoint).ScalarBaseMult(&proof.D)
	if !phi.Equal(bPlusCPk) {
		return errors.New("sk knowledge proof verification failed, second check")
	}
	return nil
}
