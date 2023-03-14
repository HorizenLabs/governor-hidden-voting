package crypto

import (
	"encoding/json"
	"errors"
	"io"
	"math/big"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

// ProofCorrectDecryption is a cryptographic proof of correct decryption of an election tally
type ProofCorrectDecryption struct {
	U arith.CurvePoint
	V arith.CurvePoint
	S arith.Scalar
}

// ProveCorrectDecryption generates a proof of correct decryption of tally.
func ProveCorrectDecryption(
	// Implementation based on https://eprint.iacr.org/2016/765.pdf, section 4.4
	reader io.Reader,
	tally *EncryptedTally,
	keyPair *KeyPair) (*ProofCorrectDecryption, error) {

	r, v, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, err
	}
	u := new(arith.CurvePoint).ScalarMult(&tally.Votes.A, r)

	bytesPk, err := keyPair.Pk.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesA, err := tally.Votes.A.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytesB, err := tally.Votes.B.MarshalBinary()
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
	proof.U.Set(u)
	proof.V.Set(v)
	proof.S.Set(s)
	return proof, nil
}

// VerifyCorrectDecryption verifies a proof of correct tally decryption.
// Parameter result is the total number of Yes votes
func VerifyCorrectDecryption(
	// Implementation based on https://eprint.iacr.org/2016/765.pdf, section 4.4
	proof *ProofCorrectDecryption,
	tally *EncryptedTally,
	result int64,
	pk *arith.CurvePoint) error {

	m, err := json.Marshal(proof)
	if err != nil {
		return err
	}
	proof = new(ProofCorrectDecryption)
	err = json.Unmarshal(m, proof)
	if err != nil {
		return err
	}

	bytesPk, err := pk.MarshalBinary()
	if err != nil {
		return err
	}
	bytesA, err := tally.Votes.A.MarshalBinary()
	if err != nil {
		return err
	}
	bytesB, err := tally.Votes.B.MarshalBinary()
	if err != nil {
		return err
	}
	bytesU, err := proof.U.MarshalBinary()
	if err != nil {
		return err
	}
	bytesV, err := proof.V.MarshalBinary()
	if err != nil {
		return err
	}
	c := arith.FiatShamirChallenge(
		bytesPk,
		bytesA,
		bytesB,
		bytesU,
		bytesV)
	d := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(-result)))
	d = new(arith.CurvePoint).Add(d, &tally.Votes.B)
	sA := new(arith.CurvePoint).ScalarMult(&tally.Votes.A, &proof.S)
	cD := new(arith.CurvePoint).ScalarMult(d, c.Scalar())
	uPlusCD := new(arith.CurvePoint).Add(&proof.U, cD)
	if !sA.Equal(uPlusCD) {
		return errors.New("decryption proof verification failed, first check")
	}

	sG := new(arith.CurvePoint).ScalarBaseMult(&proof.S)
	cPk := new(arith.CurvePoint).ScalarMult(pk, c.Scalar())
	vPlusCPk := new(arith.CurvePoint).Add(&proof.V, cPk)
	if !sG.Equal(vPlusCPk) {
		return errors.New("decryption proof verification failed, second check")
	}
	return nil
}
