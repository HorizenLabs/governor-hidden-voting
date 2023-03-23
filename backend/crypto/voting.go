package crypto

import (
	"errors"
	"io"
	"math"
	"math/big"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
)

type Vote int64

const (
	No Vote = iota
	Yes
)

// EncryptedVote represents a vote encrypted with EC-ElGamal.
// Thanks to the linear homomorphic properties of ElGamal encryption,
// it can represent both a single 0-1 vote and the sum of an arbitrary number
// of such votes.
type EncryptedVote struct {
	A arith.CurvePoint `json:"a"`
	B arith.CurvePoint `json:"b"`
}

func NewEncryptedVote() *EncryptedVote {
	vote := new(EncryptedVote)
	zero := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(0)))
	vote.A.Set(zero)
	vote.B.Set(zero)
	return vote
}

// Set sets the receiver to v and returns it.
func (e *EncryptedVote) Set(v *EncryptedVote) *EncryptedVote {
	e.A.Set(&v.A)
	e.B.Set(&v.B)
	return e
}

// Add sets the receiver to the sum of a and b and returns it.
func (e *EncryptedVote) Add(a, b *EncryptedVote) *EncryptedVote {
	e.A.Add(&a.A, &b.A)
	e.B.Add(&a.B, &b.B)
	return e
}

// Encrypt encrypts a vote and returns the encrypted vote and the secret
// random scalar used for ElGamal encryption. This scalar is useful for
// generating a proof of vote well-formedness with function ProveVoteWellFormedness.
func (vote Vote) Encrypt(reader io.Reader, pk *arith.CurvePoint) (*EncryptedVote, *arith.Scalar, error) {
	encodedVote := encode(vote)
	return encryptInternal(encodedVote, reader, pk)
}

// Decrypt decrypts an encrypted vote and returns the result.
// Parameter n should be an upper bound on the result.
//
// If the encrypted vote has been obtained by summing a number m 0-1 votes,
// then m can be used as upper bound.
func (vote *EncryptedVote) Decrypt(sk *arith.Scalar, n int64) (Vote, error) {
	encodedVote := vote.decryptInternal(sk)
	return decode(encodedVote, n)
}

// Encrypt an encodedVote using ElGamal encryption
func encryptInternal(
	encodedVote *arith.CurvePoint,
	reader io.Reader,
	pk *arith.CurvePoint) (*EncryptedVote, *arith.Scalar, error) {
	r, a, err := arith.RandomCurvePoint(reader)
	if err != nil {
		return nil, nil, err
	}
	b := new(arith.CurvePoint).ScalarMult(pk, r)
	b = new(arith.CurvePoint).Add(b, encodedVote)
	encryptedVote := new(EncryptedVote)
	encryptedVote.A.Set(a)
	encryptedVote.B.Set(b)
	return encryptedVote, r, nil
}

// DecryptInternal computes the decryption of an encrypted vote (a, b), returning
// the curve point b - sk*a.
func (vote *EncryptedVote) decryptInternal(sk *arith.Scalar) *arith.CurvePoint {
	skNegA := new(arith.CurvePoint).ScalarMult(&vote.A, new(arith.Scalar).Neg(sk))
	return new(arith.CurvePoint).Add(skNegA, &vote.B)
}

// Encode encodes a vote m as the curve point m*g.
func encode(vote Vote) *arith.CurvePoint {
	scalar := arith.NewScalar(big.NewInt(int64(vote)))
	return new(arith.CurvePoint).ScalarBaseMult(scalar)
}

// Decode decodes an encoded vote. Since this requires solving a dlog problem, an
// upper bound n on the result should be provided.
func decode(encodedVote *arith.CurvePoint, n int64) (Vote, error) {
	// solve dlog problem via baby-step giant-step algorithm
	m := int64(math.Ceil(math.Sqrt(float64(n + 1))))
	gPow := make(map[string]int64)
	for j := int64(0); j < m; j++ {
		val := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(int64(j))))
		valBinary, err := val.MarshalBinary()
		if err != nil {
			return 0, err
		}
		gPow[string(valBinary)] = j
	}
	mNegG := new(arith.CurvePoint).ScalarBaseMult(arith.NewScalar(big.NewInt(int64(-m))))
	gamma := new(arith.CurvePoint).Set(encodedVote)
	for i := int64(0); i < m; i++ {
		gammaBinary, err := gamma.MarshalBinary()
		if err != nil {
			return 0, err
		}
		j, ok := gPow[string(gammaBinary)]
		if ok {
			result := i*m + j
			return Vote(result), nil
		}
		gamma.Add(gamma, mNegG)
	}
	return 0, errors.New("error during vote decryption")
}
