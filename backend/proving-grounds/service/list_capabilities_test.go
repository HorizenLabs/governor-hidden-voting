package main

import (
	"context"
	"fmt"
	"testing"

	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
)

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

	server := newServerWithDummyCapabilities(t, capabilities)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			response, err := server.ListCapabilities(context.Background(), tc.request)
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

	server := newServerWithDummyCapabilities(t, capabilities)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			response, err := server.ListCapabilities(context.Background(), tc.request)
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

func newServerWithDummyCapabilities(t *testing.T, capabilities []*cs.Capability) *cryptographicServiceServer {
	server := newCryptographicServiceServer()
	for _, capability := range capabilities {
		cap := newDummyCapability(capability.Id, capability.Name, capability.Description, capability.NumArguments)
		if err := server.registerCapability(cap); err != nil {
			t.Fatal(err)
		}
	}
	return server
}
