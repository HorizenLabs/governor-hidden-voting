package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/HorizenLabs/e-voting-poc/backend/arith"
	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

func goNumber(v js.Value) (int64, error) {
	if err := isType(v, js.TypeNumber); err != nil {
		return 0, err
	}
	return int64(v.Int()), nil
}

func goCurvePoint(v js.Value) (*arith.CurvePoint, error) {
	keys := []string{"x", "y"}
	types := []js.Type{js.TypeString, js.TypeString}
	if err := isObject(v, keys, types); err != nil {
		return nil, err
	}

	xJSON := "\"" + v.Get("x").String() + "\""
	yJSON := "\"" + v.Get("y").String() + "\""
	pkJSON := "{\"x\":" + xJSON + ",\"y\":" + yJSON + "}"

	res := new(arith.CurvePoint)
	err := json.Unmarshal([]byte(pkJSON), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func goScalar(v js.Value) (*arith.Scalar, error) {
	if err := isType(v, js.TypeString); err != nil {
		return nil, err
	}

	sJSON := "\"" + v.String() + "\""
	res := new(arith.Scalar)
	err := json.Unmarshal([]byte(sJSON), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func goEncryptedVote(v js.Value) (*crypto.EncryptedVote, error) {
	keys := []string{"a", "b"}
	types := []js.Type{js.TypeObject, js.TypeObject}
	if err := isObject(v, keys, types); err != nil {
		return nil, err
	}

	a, err := goCurvePoint(v.Get("a"))
	if err != nil {
		return nil, NewFieldParsingError("a", err)
	}
	b, err := goCurvePoint(v.Get("b"))
	if err != nil {
		return nil, NewFieldParsingError("b", err)
	}

	res := new(crypto.EncryptedVote)
	res.A.Set(a)
	res.B.Set(b)
	return res, nil
}

func goKeyPair(v js.Value) (*crypto.KeyPair, error) {
	keys := []string{"pk", "sk"}
	types := []js.Type{js.TypeObject, js.TypeString}
	if err := isObject(v, keys, types); err != nil {
		return nil, err
	}

	pk, err := goCurvePoint(v.Get("pk"))
	if err != nil {
		return nil, NewFieldParsingError("pk", err)
	}
	sk, err := goScalar(v.Get("sk"))
	if err != nil {
		return nil, NewFieldParsingError("sk", err)
	}

	res := new(crypto.KeyPair)
	res.Pk.Set(pk)
	res.Sk.Set(sk)
	return res, nil
}
