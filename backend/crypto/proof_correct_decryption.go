package crypto

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

// ProofCorrectDecryption is a cryptographic proof of correct decryption of an encrypted vote
type ProofCorrectDecryption struct {
	S arith.Scalar    `json:"s"`
	C arith.Challenge `json:"c"`
}

// Set sets p to q and returns q.
func (q *ProofCorrectDecryption) Set(p *ProofCorrectDecryption) *ProofCorrectDecryption {
	q.S.Set(&p.S)
	q.C.Set(&p.C)
	return q
}

// ProveCorrectDecryption generates a proof of correct decryption of an encrypted vote.
func ProveCorrectDecryption(
	reader io.Reader,
	encryptedVote *EncryptedVote,
	keyPair *KeyPair) (*ProofCorrectDecryption, error) {
	// Implementation based on https://eprint.iacr.org/2016/765.pdf, section 4.4
	r, v, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}

	u := new(arith.CurvePoint).ScalarMult(&encryptedVote.A, r)

	bytesPk, err := keyPair.Pk.MarshalBinary()
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
	bytesU, err := u.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesV, err := v.MarshalBinary()
	if err != nil {
		return nil, err
	}
	c := arith.FiatShamirChallenge(
		bytesPk,
		bytesA,
		bytesB,
		bytesU,
		bytesV)

	s := new(arith.Scalar).Mul(c.Scalar(), &keyPair.Sk)
	s = new(arith.Scalar).Add(r, s)
	proof := new(ProofCorrectDecryption)
	proof.S.Set(s)
	proof.C.Set(c)
	return proof, nil
}

// VerifyCorrectDecryption verifies a proof of correct vote decryption.
func VerifyCorrectDecryption(
	proof *ProofCorrectDecryption,
	encryptedVote *EncryptedVote,
	decryptedVote Vote,
	pk *arith.CurvePoint) error {
	minusEncodedVote := new(arith.CurvePoint).Neg(encode(decryptedVote))
	d := new(arith.CurvePoint).Add(&encryptedVote.B, minusEncodedVote)
	return verifyCorrectDecryptionInternal(proof, encryptedVote, d, pk)
}

func verifyCorrectDecryptionInternal(proof *ProofCorrectDecryption,
	encryptedVote *EncryptedVote,
	d *arith.CurvePoint,
	pk *arith.CurvePoint) error {
	// Implementation based on https://eprint.iacr.org/2016/765.pdf, section 4.4
	m, err := json.Marshal(proof)
	if err != nil {
		return err
	}
	proof = new(ProofCorrectDecryption)
	err = json.Unmarshal(m, proof)
	if err != nil {
		return err
	}

	sA := new(arith.CurvePoint).ScalarMult(&encryptedVote.A, &proof.S)
	cD := new(arith.CurvePoint).ScalarMult(d, proof.C.Scalar())
	u := new(arith.CurvePoint).Add(sA, new(arith.CurvePoint).Neg(cD))

	sG := new(arith.CurvePoint).ScalarBaseMult(&proof.S)
	cPk := new(arith.CurvePoint).ScalarMult(pk, proof.C.Scalar())
	v := new(arith.CurvePoint).Add(sG, new(arith.CurvePoint).Neg(cPk))

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA, err := encryptedVote.A.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB, err := encryptedVote.B.MarshalBinary()
	if err != nil {
		return err
	}
	bytesU, err := u.MarshalBinary()
	if err != nil {
		return err
	}
	bytesV, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	c := arith.FiatShamirChallenge(
		bytesPk,
		bytesA,
		bytesB,
		bytesU,
		bytesV)

	if !c.Equal(&proof.C) {
		return errors.New("decryption proof verification failed")
	}
	return nil
}
