package crypto_service

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"

	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
	"golang.org/x/net/nettest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRegisterCapabilities(t *testing.T) {
	foo := newFoo()
	bar := newBar()
	service := NewCryptographicService()
	if err := service.RegisterCapability(foo); err != nil {
		t.Fatal(err)
	}
	if err := service.RegisterCapability(bar); err != nil {
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
	service := NewCryptographicService()
	if err := service.RegisterCapability(foo); err != nil {
		t.Fatal(err)
	}
	err := service.RegisterCapability(bar)
	if err == nil {
		t.Fatalf("registering a capability id twice should fail")
	}
}

func TestComputeNonExistingCapability(t *testing.T) {
	foo := newFoo()
	service := NewCryptographicService()
	if err := service.RegisterCapability(foo); err != nil {
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
	service := NewCryptographicService()
	if err := service.RegisterCapability(foo); err != nil {
		t.Fatal(err)
	}
	if err := service.RegisterCapability(bar); err != nil {
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

func TestServe(t *testing.T) {
	foo := newFoo()
	foo.shouldSucceed = true

	lis, err := nettest.NewLocalListener("tcp")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		service := NewCryptographicService()
		service.RegisterCapability(foo)
		err = service.Serve(lis)
		if err != nil {
			log.Fatal(err)
		}
	}()

	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(lis.Addr().String(), dialOpts...)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := cs.NewCryptographicServiceClient(conn)
	computeRequest := &cs.Execute{
		RequestId:    "request_id",
		CapabilityId: "foo_id",
		Arguments:    [][]byte{[]byte("Hello")},
	}
	out, err := client.Compute(context.Background(), computeRequest)

	if err != nil {
		t.Fatal(err)
	}
	if !out.GetOk() {
		t.Fatalf(out.GetError())
	}
	if foo.getNumCalls() != 1 {
		t.Fatalf("capability should have been called")
	}
}

func TestFilterCapabilities(t *testing.T) {
	capabilities := []*cs.Capability{
		{Id: "id_0", Name: "name_0", Description: "description 0", NumArguments: 0},
		{Id: "id_1", Name: "name_1", Description: "description 1", NumArguments: 1},
		{Id: "id_2", Name: "name_2", Description: "description 2", NumArguments: 2},
		{Id: "id_3", Name: "name_0", Description: "description 3", NumArguments: 3},
		{Id: "id_4", Name: "name_4", Description: "description 4", NumArguments: 3},
	}

	tests := listCapabilitiesTest{
		"NoFiltering": {
			request: &cs.ListCapabilitiesRequest{},
			out:     capabilities,
		},
		"NonExistingByName": {
			request: &cs.ListCapabilitiesRequest{
				FilterName: createString("name_non_existing"),
			},
			out: []*cs.Capability{},
		},
		"ExistingSingleByName": {
			request: &cs.ListCapabilitiesRequest{
				FilterName: createString("name_1"),
			},
			out: capabilities[1:2],
		},
		"ExistingMultipleByName": {
			request: &cs.ListCapabilitiesRequest{
				FilterName: createString("name_0"),
			},
			out: []*cs.Capability{capabilities[0], capabilities[3]},
		},
		"NonExistingByNumArgs": {
			request: &cs.ListCapabilitiesRequest{
				FilterArgs: createUint32(1000),
			},
			out: []*cs.Capability{},
		},
		"ExistingSingleByNumArgs": {
			request: &cs.ListCapabilitiesRequest{
				FilterArgs: createUint32(1),
			},
			out: capabilities[1:2],
		},
		"ExistingMultipleByNumArgs": {
			request: &cs.ListCapabilitiesRequest{
				FilterArgs: createUint32(3),
			},
			out: []*cs.Capability{capabilities[3], capabilities[4]},
		},
		"NonExistingById": {
			request: &cs.ListCapabilitiesRequest{
				FilterId: createString("id_non_existing"),
			},
			out: []*cs.Capability{},
		},
		"ExistingById": {
			request: &cs.ListCapabilitiesRequest{
				FilterId: createString("id_2"),
			},
			out: capabilities[2:3],
		},
		"CompatibleIdNameArgs": {
			request: &cs.ListCapabilitiesRequest{
				FilterId:   createString("id_2"),
				FilterName: createString("name_2"),
				FilterArgs: createUint32(2),
			},
			out: capabilities[2:3],
		},
		"IncompatibleIdName": {
			request: &cs.ListCapabilitiesRequest{
				FilterId:   createString("id_1"),
				FilterName: createString("name_2"),
			},
			out: []*cs.Capability{},
		},
		"IncompatibleIdArgs": {
			request: &cs.ListCapabilitiesRequest{
				FilterId:   createString("id_1"),
				FilterArgs: createUint32(2),
			},
			out: []*cs.Capability{},
		},
		"IncompatibleNameArgs": {
			request: &cs.ListCapabilitiesRequest{
				FilterName: createString("id_1"),
				FilterArgs: createUint32(2),
			},
			out: []*cs.Capability{},
		},
	}

	service := newServiceWithDummyCapabilities(t, capabilities)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			response, err := service.ListCapabilities(context.Background(), tc.request)
			if err != nil {
				t.Fatal(err)
			}
			if len(response.GetCapabilities()) != len(tc.out) {
				t.Fatalf("wrong number of capabilities: expected %d, got %d", len(tc.out), len(response.GetCapabilities()))
			}
			for i, cap := range response.GetCapabilities() {
				if cap.GetId() != tc.out[i].Id {
					t.Fatalf("wrong %d-th capability: expected id \"%s\", got id \"%s\"", i, tc.out[i].Id, cap.GetId())
				}
			}
		})
	}
}

func TestPaginateCapabilities(t *testing.T) {
	numCapabilities := 123
	capabilities := []*cs.Capability{}
	for i := 0; i < numCapabilities; i++ {
		capability := &cs.Capability{
			Id: fmt.Sprintf("id_%3d", i),
		}
		capabilities = append(capabilities, capability)
	}

	tests := listCapabilitiesTest{
		"NoPagination": {
			request: &cs.ListCapabilitiesRequest{},
			out:     capabilities,
		},
		"DefaultPageSize/Page0": {
			request: &cs.ListCapabilitiesRequest{
				PageNum: createUint32(0),
			},
			out: capabilities[0:50],
		},
		"DefaultPageSize/Page1": {
			request: &cs.ListCapabilitiesRequest{
				PageNum: createUint32(1),
			},
			out: capabilities[50:100],
		},
		"DefaultPageSize/LastPage": {
			request: &cs.ListCapabilitiesRequest{
				PageNum: createUint32(2),
			},
			out: capabilities[100:],
		},
		"CustomPageSize/Page0": {
			request: &cs.ListCapabilitiesRequest{
				PageSize: createUint32(5),
				PageNum:  createUint32(0),
			},
			out: capabilities[0:5],
		},
		"CustomPageSize/Page1": {
			request: &cs.ListCapabilitiesRequest{
				PageSize: createUint32(5),
				PageNum:  createUint32(1),
			},
			out: capabilities[5:10],
		},
		"CustomPageSize/LastPage": {
			request: &cs.ListCapabilitiesRequest{
				PageSize: createUint32(6),
				PageNum:  createUint32(20),
			},
			out: capabilities[120:],
		},
	}

	service := newServiceWithDummyCapabilities(t, capabilities)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			response, err := service.ListCapabilities(context.Background(), tc.request)
			if err != nil {
				t.Fatal(err)
			}
			if len(response.GetCapabilities()) != len(tc.out) {
				t.Fatalf("wrong number of capabilities: expected %d, got %d", len(tc.out), len(response.GetCapabilities()))
			}
			for i, cap := range response.GetCapabilities() {
				if cap.GetId() != tc.out[i].Id {
					t.Fatalf("wrong %d-th capability: expected id \"%s\", got id \"%s\"", i, tc.out[i].Id, cap.GetId())
				}
			}
		})
	}
}

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
	service := NewCryptographicService()
	if err := service.RegisterCapability(foo); err != nil {
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

type listCapabilitiesTest map[string]struct {
	request *cs.ListCapabilitiesRequest
	out     []*cs.Capability
}

func createUint32(x uint32) *uint32 {
	return &x
}

func createString(s string) *string {
	return &s
}

func newServiceWithDummyCapabilities(t *testing.T, capabilities []*cs.Capability) *CryptographicService {
	service := NewCryptographicService()
	for _, capability := range capabilities {
		cap := newDummyCapability(capability.Id, capability.Name, capability.Description, capability.NumArguments)
		if err := service.RegisterCapability(cap); err != nil {
			t.Fatal(err)
		}
	}
	return service
}
