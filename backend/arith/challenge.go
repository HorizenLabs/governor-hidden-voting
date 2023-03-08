package arith

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

const NumBytesChallenge = 128 / 8

type Challenge struct {
	val big.Int
}

func (c *Challenge) String() string {
	return c.val.String()
}

func FiatShamirChallenge(data ...[]byte) *Challenge {
	hashBytes := crypto.Keccak256(data...)
	// We employ 128 bits challenges
	return &Challenge{*new(big.Int).SetBytes(hashBytes[16:])}
}

func RandomChallenge(reader io.Reader) (*Challenge, error) {
	val, err := rand.Int(reader, challengeModulus())
	return &Challenge{*val}, err
}

func challengeModulus() *big.Int {
	// We employ 128 bits challenges
	return new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)
}

func (c *Challenge) Scalar() *Scalar {
	return NewScalar(&c.val)
}

func (e *Challenge) Add(a, b *Challenge) *Challenge {
	e.val.Set(new(big.Int).Add(&a.val, &b.val))
	e.val.Mod(&e.val, challengeModulus())
	return e
}

func (e *Challenge) Sub(a, b *Challenge) *Challenge {
	e.val.Set(new(big.Int).Sub(&a.val, &b.val))
	e.val.Mod(&e.val, challengeModulus())
	return e
}

func (e *Challenge) Mul(a, b *Challenge) *Challenge {
	e.val.Set(new(big.Int).Mul(&a.val, &b.val))
	e.val.Mod(&e.val, challengeModulus())
	return e
}

func (e *Challenge) Neg(a *Challenge) *Challenge {
	e.val.Set(new(big.Int).Sub(challengeModulus(), &a.val))
	return e
}

func (a *Challenge) Equal(b *Challenge) bool {
	return a.val.Cmp(&b.val) == 0
}

func (e *Challenge) Marshal() []byte {
	buf := make([]byte, NumBytesChallenge)
	e.val.FillBytes(buf)
	return buf
}

func (e *Challenge) Unmarshal(m []byte) ([]byte, error) {
	if len(m) < NumBytesChallenge {
		return nil, errors.New("message too short")
	}
	ret := new(big.Int).SetBytes(m[:NumBytesChallenge])
	e.val.Set(ret)
	return m[NumBytesChallenge:], nil
}
