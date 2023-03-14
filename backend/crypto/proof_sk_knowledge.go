package crypto

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

// ProofSkKnowledge is a cryptographic proof of knowledge of the secret key of an ElGamal KeyPair
type ProofSkKnowledge struct {
	B arith.CurvePoint
	D arith.Scalar
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
	proof.B.Set(b)
	proof.D.Set(d)
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

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB, err := proof.B.MarshalBinary()
	if err != nil {
		return err
	}
	c := arith.FiatShamirChallenge(bytesPk, bytesB)
	cPk := new(arith.CurvePoint).ScalarMult(pk, c.Scalar())
	bPlusCPk := new(arith.CurvePoint).Add(&proof.B, cPk)
	phi := new(arith.CurvePoint).ScalarBaseMult(&proof.D)
	if !phi.Equal(bPlusCPk) {
		return errors.New("sk knowledge proof verification failed")
	}
	return nil
}
