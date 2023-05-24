package main

import (
	"crypto/rand"
	"fmt"
	"syscall/js"

	"github.com/HorizenLabs/e-voting-poc/backend/crypto"
)

func main() {
	js.Global().Set("goNewKeyPairWithProof", promiseWrapper(newKeyPairWithProof))
	js.Global().Set("goEncryptVoteWithProof", promiseWrapper(encryptVoteWithProof))
	js.Global().Set("goDecryptTallyWithProof", promiseWrapper(decryptTallyWithProof))
	js.Global().Set("goAddEncryptedVotes", promiseWrapper(addEncryptedVotes))
	<-make(chan bool)
}

func newKeyPairWithProof(this js.Value, args []js.Value) (js.Value, error) {
	if err := checkArgsNum(args, 0); err != nil {
		return js.Null(), err
	}

	keyPair, proof, err := crypto.NewKeyPairWithProof(rand.Reader)
	if err != nil {
		return js.Null(), err
	}

	jsKeyPair, err := jsValueKeyPair(keyPair)
	if err != nil {
		return js.Null(), err
	}

	jsProof, err := jsValueProofSkKnowledge(proof)
	if err != nil {
		return js.Null(), err
	}

	result := jsObject{
		"keyPair": jsKeyPair,
		"proof":   jsProof,
	}
	return js.ValueOf(result), nil
}

func encryptVoteWithProof(this js.Value, args []js.Value) (js.Value, error) {
	if err := checkArgsNum(args, 2); err != nil {
		return js.Null(), err
	}
	vote, err := goNumber(args[0])
	if err != nil {
		return js.Null(), NewArgParsingError(0, err)
	}
	pk, err := goCurvePoint(args[1])
	if err != nil {
		return js.Null(), NewArgParsingError(1, err)
	}

	encryptedVote, proof, err := crypto.EncryptVoteWithProof(rand.Reader, vote, pk)
	if err != nil {
		return js.Null(), err
	}

	jsEncryptedVote, err := jsValueEncryptedVote(encryptedVote)
	if err != nil {
		return js.Null(), err
	}
	jsProof, err := jsValueProofVoteWellFormedness(proof)
	if err != nil {
		return js.Null(), err
	}

	result := jsObject{
		"encryptedVote": jsEncryptedVote,
		"proof":         jsProof,
	}
	return js.ValueOf(result), nil
}

func decryptTallyWithProof(this js.Value, args []js.Value) (js.Value, error) {
	if err := checkArgsNum(args, 3); err != nil {
		return js.Null(), err
	}
	tally, err := goEncryptedVote(args[0])
	if err != nil {
		return js.Null(), NewArgParsingError(0, err)
	}
	n, err := goNumber(args[1])
	if err != nil {
		return js.Null(), NewArgParsingError(1, err)
	}
	keyPair, err := goKeyPair(args[2])
	if err != nil {
		return js.Null(), NewArgParsingError(2, err)
	}

	jsDecryptedTally, proof, err := crypto.DecryptTallyWithProof(rand.Reader, tally, n, keyPair)
	if err != nil {
		return js.Null(), err
	}

	jsProof, err := jsValueProofCorrectDecryption(proof)
	if err != nil {
		return js.Null(), err
	}

	result := jsObject{
		"result": jsDecryptedTally,
		"proof":  jsProof,
	}
	return js.ValueOf(result), nil
}

func addEncryptedVotes(this js.Value, args []js.Value) (js.Value, error) {
	if err := checkArgsNum(args, 2); err != nil {
		return js.Null(), err
	}
	vote0, err := goEncryptedVote(args[0])
	if err != nil {
		return js.Null(), NewArgParsingError(0, err)
	}
	vote1, err := goEncryptedVote(args[1])
	if err != nil {
		return js.Null(), NewArgParsingError(1, err)
	}

	vote := new(crypto.EncryptedVote).Add(vote0, vote1)

	jsVote, err := jsValueEncryptedVote(vote)
	if err != nil {
		return js.Null(), err
	}

	return jsVote, nil
}

func checkArgsNum(args []js.Value, num int) error {
	if len(args) != num {
		return fmt.Errorf("function takes %d arguments", num)
	}
	return nil
}

func promiseWrapper(f func(js.Value, []js.Value) (js.Value, error)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		handler := js.FuncOf(func(handlerThis js.Value, handlerArgs []js.Value) interface{} {
			resolve := handlerArgs[0]
			reject := handlerArgs[1]

			go func() {
				data, err := f(this, args)
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
				} else {
					resolve.Invoke(js.ValueOf(data))
				}
			}()

			return nil
		})

		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

type argParsingError struct {
	argNum int
	err    error
}

func NewArgParsingError(argNum int, err error) *argParsingError {
	return &argParsingError{
		argNum: argNum,
		err:    err,
	}
}

func (e *argParsingError) Error() string {
	return fmt.Sprintf("error parsing argument %d: %v", e.argNum, e.err)
}
