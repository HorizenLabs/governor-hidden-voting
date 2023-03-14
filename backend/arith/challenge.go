package arith

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
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
	challenge := new(Challenge)
	challenge.val.SetBytes(hashBytes[16:])
	return challenge
}

func RandomChallenge(reader io.Reader) (*Challenge, error) {
	val, err := rand.Int(reader, challengeModulus())
	if err != nil {
		return nil, err
	}
	challenge := new(Challenge)
	challenge.val.Set(val)
	return challenge, err
}

func challengeModulus() *big.Int {
	// We employ 128 bits challenges
	return new(big.Int).Lsh(big.NewInt(1), 128)
}

func (c *Challenge) Scalar() *Scalar {
	return NewScalar(&c.val)
}

func (e *Challenge) Set(a *Challenge) *Challenge {
	e.val.Set(&a.val)
	return e
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

func (e *Challenge) MarshalBinary() ([]byte, error) {
	buf := make([]byte, NumBytesChallenge)
	e.val.FillBytes(buf)
	return buf, nil
}

func (e *Challenge) UnmarshalBinary(m []byte) error {
	if len(m) != NumBytesChallenge {
		return fmt.Errorf("challenge should be represented with %d bytes", NumBytesScalar)
	}
	e.val.Set(new(big.Int).SetBytes(m))
	return nil
}

func (e Challenge) MarshalJSON() ([]byte, error) {
	bytesE, err := e.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return json.Marshal(bytesE)
}

func (e *Challenge) UnmarshalJSON(data []byte) error {
	var bytesE []byte
	err := json.Unmarshal(data, &bytesE)
	if err != nil {
		return err
	}
	return e.UnmarshalBinary(bytesE)
}
