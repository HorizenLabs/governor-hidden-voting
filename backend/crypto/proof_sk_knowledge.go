package crypto

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

// ProofSkKnowledge is a cryptographic proof of knowledge of the secret key of an ElGamal KeyPair
type ProofSkKnowledge struct {
	S arith.Scalar    `json:"s"`
	C arith.Challenge `json:"c"`
}

// ProveSkKnowledge generates a proof of knowledge of the secret key of an ElGamal KeyPair
func ProveSkKnowledge(reader io.Reader, keyPair *KeyPair) (*ProofSkKnowledge, error) {
	// Implementation based on https://eprint.iacr.org/2016/765.pdf, section 4.3
	r, v, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}

	bytesPk, err := keyPair.Pk.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesV, err := v.MarshalBinary()
	if err != nil {
		return nil, err
	}
	c := arith.FiatShamirChallenge(bytesPk, bytesV)

	s := new(arith.Scalar).Mul(c.Scalar(), &keyPair.Sk)
	s = new(arith.Scalar).Add(r, s)
	proof := new(ProofSkKnowledge)
	proof.S.Set(s)
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
	v := new(arith.CurvePoint).ScalarBaseMult(&proof.S)
	v.Add(v, new(arith.CurvePoint).Neg(cPk))

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return err
	}
	bytesV, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	c := arith.FiatShamirChallenge(bytesPk, bytesV)

	if !c.Equal(&proof.C) {
		return errors.New("sk knowledge proof verification failed")
	}
	return nil
}
