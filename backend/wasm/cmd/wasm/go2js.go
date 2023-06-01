package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

type jsObject = map[string]interface{}

func jsValueCurvePoint(p *arith.CurvePoint) (js.Value, error) {
	pJSON, err := json.Marshal(p)
	if err != nil {
		return js.Null(), err
	}
	pInternal := new(arith.CurvePointInternal)
	err = json.Unmarshal(pJSON, pInternal)
	if err != nil {
		return js.Null(), err
	}
	return js.ValueOf(jsObject{
		"x": pInternal.X.String(),
		"y": pInternal.Y.String(),
	}), nil
}

func jsValueScalar(s *arith.Scalar) (js.Value, error) {
	sJSON, err := json.Marshal(s)
	if err != nil {
		return js.Null(), err
	}
	return js.ValueOf(string(sJSON)), nil
}

func jsValueChallenge(c *arith.Challenge) (js.Value, error) {
	cJSON, err := json.Marshal(c)
	if err != nil {
		return js.Null(), err
	}
	return js.ValueOf(string(cJSON)), nil
}

func jsValueKeyPair(keyPair *crypto.KeyPair) (js.Value, error) {
	pk, err := jsValueCurvePoint(&keyPair.Pk)
	if err != nil {
		return js.Null(), err
	}
	sk, err := jsValueScalar(&keyPair.Sk)
	if err != nil {
		return js.Null(), err
	}
	result := jsObject{
		"pk": pk,
		"sk": sk,
	}
	return js.ValueOf(result), nil
}

func jsValueEncryptedVote(encryptedVote *crypto.EncryptedVote) (js.Value, error) {
	a, err := jsValueCurvePoint(&encryptedVote.A)
	if err != nil {
		return js.Null(), err
	}
	b, err := jsValueCurvePoint(&encryptedVote.B)
	if err != nil {
		return js.Null(), err
	}
	result := jsObject{
		"a": a,
		"b": b,
	}
	return js.ValueOf(result), err
}

func jsValueProofSkKnowledge(proof *crypto.ProofSkKnowledge) (js.Value, error) {
	s, err := jsValueScalar(&proof.S)
	if err != nil {
		return js.Null(), err
	}
	c, err := jsValueChallenge(&proof.C)
	if err != nil {
		return js.Null(), err
	}
	result := jsObject{
		"s": s,
		"c": c,
	}
	return js.ValueOf(result), nil
}

func jsValueProofVoteWellFormedness(proof *crypto.ProofVoteWellFormedness) (js.Value, error) {
	r0, err := jsValueScalar(&proof.R0)
	if err != nil {
		return js.Null(), err
	}
	r1, err := jsValueScalar(&proof.R1)
	if err != nil {
		return js.Null(), err
	}
	c0, err := jsValueChallenge(&proof.C0)
	if err != nil {
		return js.Null(), err
	}
	c1, err := jsValueChallenge(&proof.C1)
	if err != nil {
		return js.Null(), err
	}
	result := jsObject{
		"r0": r0,
		"r1": r1,
		"c0": c0,
		"c1": c1,
	}
	return js.ValueOf(result), nil
}

func jsValueProofCorrectDecryption(proof *crypto.ProofCorrectDecryption) (js.Value, error) {
	s, err := jsValueScalar(&proof.S)
	if err != nil {
		return js.Null(), err
	}
	c, err := jsValueChallenge(&proof.C)
	if err != nil {
		return js.Null(), err
	}
	result := jsObject{
		"s": s,
		"c": c,
	}
	return js.ValueOf(result), nil
}
