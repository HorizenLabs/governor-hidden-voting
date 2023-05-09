package main

import (
	"context"
	"reflect"
	"testing"

	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
)

func newFoo() *dummyCapability {
	return newDummyCapability(
		"foo_id",
		"foo",
		"foo description",
		0,
	)
}

func newBar() *dummyCapability {
	return newDummyCapability(
		"bar_id",
		"bar",
		"bar description",
		0,
	)
}

func capabilityComputedFixture(
	t *testing.T,
	requestId string,
	arguments [][]byte,
	success bool,
) (*dummyCapability, *cs.Result) {
	foo := newFoo()
	foo.shouldSucceed = success
	service := newCryptographicServiceServer()
	if err := service.registerCapability(foo); err != nil {
		t.Fatal(err)
	}
	request := &cs.Execute{
		RequestId:    requestId,
		CapabilityId: foo.Id(),
		Arguments:    arguments,
	}
	response, err := service.Compute(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	return foo, response
}

func TestRegisterCapabilities(t *testing.T) {
	foo := newFoo()
	bar := newBar()
	service := newCryptographicServiceServer()
	if err := service.registerCapability(foo); err != nil {
		t.Fatal(err)
	}
	if err := service.registerCapability(bar); err != nil {
		t.Fatal(err)
	}

	if len(service.capabilities) != 2 {
		t.Fatalf("len(service.capabilities): expected %d, got %d", 2, len(service.capabilities))
	}
}

func TestRegisterCapabilityIdTwice(t *testing.T) {
	foo := newFoo()
	bar := newBar()
	bar.id = foo.id
	service := newCryptographicServiceServer()
	if err := service.registerCapability(foo); err != nil {
		t.Fatal(err)
	}
	err := service.registerCapability(bar)
	if err == nil {
		t.Fatalf("registering a capability id twice should fail")
	}
}

func TestComputeNonExistingCapability(t *testing.T) {
	foo := newFoo()
	service := newCryptographicServiceServer()
	if err := service.registerCapability(foo); err != nil {
		t.Fatal(err)
	}

	request := &cs.Execute{
		RequestId:    "0",
		CapabilityId: "invalid_id",
		Arguments:    [][]byte{},
	}

	response, err := service.Compute(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if response.GetError() == "" {
		t.Fatalf("compute with unregistered capabilityId should error")
	}
}

func TestCorrectComputeCallRouting(t *testing.T) {
	foo := newFoo()
	bar := newBar()
	service := newCryptographicServiceServer()
	if err := service.registerCapability(foo); err != nil {
		t.Fatal(err)
	}
	if err := service.registerCapability(bar); err != nil {
		t.Fatal(err)
	}

	// Sanity checks
	if foo.getNumCalls() != 0 {
		t.Fatalf("foo.numCalls(): expected %d, got %d", 0, foo.getNumCalls())
	}
	if bar.getNumCalls() != 0 {
		t.Fatalf("bar.numCalls(): expected %d, got %d", 0, bar.getNumCalls())
	}

	// Invoke foo
	requestFoo := &cs.Execute{
		RequestId:    "0",
		CapabilityId: "foo_id",
		Arguments:    [][]byte{},
	}
	_, err := service.Compute(context.Background(), requestFoo)
	if err != nil {
		t.Fatal(err)
	}
	if foo.getNumCalls() != 1 {
		t.Fatalf("foo.numCalls(): expected %d, got %d", 1, foo.getNumCalls())
	}
	if bar.getNumCalls() != 0 {
		t.Fatalf("bar.numCalls(): expected %d, got %d", 0, bar.getNumCalls())
	}

	// Invoke bar
	requestBar := &cs.Execute{
		RequestId:    "1",
		CapabilityId: "bar_id",
		Arguments:    [][]byte{},
	}
	_, err = service.Compute(context.Background(), requestBar)
	if err != nil {
		t.Fatal(err)
	}
	if foo.getNumCalls() != 1 {
		t.Fatalf("foo.numCalls(): expected %d, got %d", 1, foo.getNumCalls())
	}
	if bar.getNumCalls() != 1 {
		t.Fatalf("bar.numCalls(): expected %d, got %d", 1, bar.getNumCalls())
	}
}

func TestCorrectArgumentsPassedToCompute(t *testing.T) {
	requestId := "id0"
	arguments := [][]byte{[]byte("hello"), []byte("world")}
	success := true
	capability, _ := capabilityComputedFixture(t, requestId, arguments, success)

	if !reflect.DeepEqual(arguments, capability.computeCallsLog[0].in) {
		t.Fatalf("arguments incorrectly passed")
	}
}

func TestCorrectRequestIdWhenComputeSucceeds(t *testing.T) {
	requestId := "id0"
	arguments := [][]byte{}
	success := true
	_, response := capabilityComputedFixture(t, requestId, arguments, success)

	if response.GetRequestId() != requestId {
		t.Fatalf("RequestId: expected %s, got %s", requestId, response.GetRequestId())
	}
}

func TestEmptyErrorWhenComputeSucceeds(t *testing.T) {
	requestId := "id0"
	arguments := [][]byte{}
	success := true
	_, response := capabilityComputedFixture(t, requestId, arguments, success)

	if response.GetError() != "" {
		t.Fatalf("Error: expected \"%s\", got \"%s\"", "", response.GetError())
	}
}

func TestOkTrueWhenComputeSucceeds(t *testing.T) {
	requestId := "id0"
	arguments := [][]byte{}
	success := true
	_, response := capabilityComputedFixture(t, requestId, arguments, success)

	if !response.GetOk() {
		t.Fatalf("Ok: expected %t, got %t", true, response.GetOk())
	}
}

func TestCorrectResultWhenComputeSucceeds(t *testing.T) {
	requestId := "id0"
	arguments := [][]byte{}
	success := true
	capability, response := capabilityComputedFixture(t, requestId, arguments, success)

	if !reflect.DeepEqual(capability.getResult(), response.GetResult()) {
		t.Fatalf("Result: incorrect")
	}
}

func TestCorrectRequestIdWhenComputeFails(t *testing.T) {
	requestId := "id0"
	arguments := [][]byte{}
	success := false
	_, response := capabilityComputedFixture(t, requestId, arguments, success)

	if response.GetRequestId() != requestId {
		t.Fatalf("RequestId: expected %s, got %s", requestId, response.GetRequestId())
	}
}

func TestCorrectErrorMessageWhenComputeFails(t *testing.T) {
	requestId := "id0"
	arguments := [][]byte{}
	success := false
	capability, response := capabilityComputedFixture(t, requestId, arguments, success)

	if response.GetError() != capability.getErrorMessage() {
		t.Fatalf("Error: expected \"%s\", got \"%s\"", capability.getErrorMessage(), response.GetError())
	}
}

func TestOkFalseWhenComputeFails(t *testing.T) {
	requestId := "id0"
	arguments := [][]byte{}
	success := false
	_, response := capabilityComputedFixture(t, requestId, arguments, success)

	if response.GetOk() {
		t.Fatalf("Ok: expected %t, got %t", false, response.GetOk())
	}
}

func TestEmptyResultWhenComputeFails(t *testing.T) {
	requestId := "id0"
	arguments := [][]byte{}
	success := false
	_, response := capabilityComputedFixture(t, requestId, arguments, success)

	if response.GetResult() != nil {
		t.Fatalf("Result: expected empty")
	}
}
