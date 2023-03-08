package arith

import (
	"crypto/rand"
	"testing"
)

func TestUnmarshalChallengeMessageTooShort(t *testing.T) {
	a, err := RandomChallenge(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	m := a.Marshal()
	_, err = a.Unmarshal(m[1:])
	if err == nil {
		t.Fatalf("should be impossible to unmarshal a message shorter than %d bytes into a challenge", NumBytesScalar)
	}
}
