package main

import (
	"context"
	"testing"

	"github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/capabilities"
	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
	"github.com/google/uuid"
)

type listCapabilitiesTests map[string]struct {
	filterName           *string
	filterId             *string
	filterArgs           *uint32
	pageNum              *uint32
	pageSize             *uint32
	expectedCapabilities []*cs.Capability
}

func TestListCapabilities(t *testing.T) {
	createUint32 := func(x uint32) *uint32 {
		return &x
	}
	tests := listCapabilitiesTests{
		"AllCapabilities": {
			expectedCapabilities: capabilities.ListCapabilities(),
		},
		"FilterName/new_key_pair_with_proof": {
			filterName:           &capabilities.Capabilities[capabilities.NEW_KEY_PAIR_WITH_PROOF_ID].Name,
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.NEW_KEY_PAIR_WITH_PROOF_ID]},
		},
		"FilterName/encrypt_vote_with_proof": {
			filterName:           &capabilities.Capabilities[capabilities.ENCRYPT_VOTE_WITH_PROOF_ID].Name,
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.ENCRYPT_VOTE_WITH_PROOF_ID]},
		},
		"FilterName/decrypt_tally_with_proof": {
			filterName:           &capabilities.Capabilities[capabilities.DECRYPT_TALLY_WITH_PROOF_ID].Name,
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.DECRYPT_TALLY_WITH_PROOF_ID]},
		},
		"FilterId/new_key_pair_with_proof": {
			filterId:             &capabilities.Capabilities[capabilities.NEW_KEY_PAIR_WITH_PROOF_ID].Id,
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.NEW_KEY_PAIR_WITH_PROOF_ID]},
		},
		"FilterId/encrypt_vote_with_proof": {
			filterId:             &capabilities.Capabilities[capabilities.ENCRYPT_VOTE_WITH_PROOF_ID].Id,
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.ENCRYPT_VOTE_WITH_PROOF_ID]},
		},
		"FilterId/decrypt_tally_with_proof": {
			filterId:             &capabilities.Capabilities[capabilities.DECRYPT_TALLY_WITH_PROOF_ID].Id,
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.DECRYPT_TALLY_WITH_PROOF_ID]},
		},
		"FilterArgs/0": {
			filterArgs:           createUint32(0),
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.NEW_KEY_PAIR_WITH_PROOF_ID]},
		},
		"FilterArgs/1": {
			filterArgs:           createUint32(1),
			expectedCapabilities: []*cs.Capability{},
		},
		"FilterArgs/2": {
			filterArgs:           createUint32(2),
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.ENCRYPT_VOTE_WITH_PROOF_ID]},
		},
		"FilterArgs/3": {
			filterArgs:           createUint32(3),
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.DECRYPT_TALLY_WITH_PROOF_ID]},
		},
		"FilterArgs/4": {
			filterArgs:           createUint32(4),
			expectedCapabilities: []*cs.Capability{},
		},
		"PageSize/0/PageNum/0": {
			pageNum:              createUint32(0),
			pageSize:             createUint32(0),
			expectedCapabilities: []*cs.Capability{},
		},
		"PageSize/1/PageNum/0": {
			pageNum:              createUint32(0),
			pageSize:             createUint32(1),
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.ENCRYPT_VOTE_WITH_PROOF_ID]},
		},
		"PageSize/1/PageNum/1": {
			pageNum:              createUint32(1),
			pageSize:             createUint32(1),
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.NEW_KEY_PAIR_WITH_PROOF_ID]},
		},
		"PageSize/1/PageNum/2": {
			pageNum:              createUint32(2),
			pageSize:             createUint32(1),
			expectedCapabilities: []*cs.Capability{capabilities.Capabilities[capabilities.DECRYPT_TALLY_WITH_PROOF_ID]},
		},
		"PageSize/1/PageNum/3": {
			pageNum:              createUint32(3),
			pageSize:             createUint32(1),
			expectedCapabilities: []*cs.Capability{},
		},
		"PageSize/2/PageNum/0": {
			pageNum:  createUint32(0),
			pageSize: createUint32(2),
			expectedCapabilities: []*cs.Capability{
				capabilities.Capabilities[capabilities.ENCRYPT_VOTE_WITH_PROOF_ID],
				capabilities.Capabilities[capabilities.NEW_KEY_PAIR_WITH_PROOF_ID],
			},
		},
		"PageSize/2/PageNum/1": {
			pageNum:  createUint32(1),
			pageSize: createUint32(2),
			expectedCapabilities: []*cs.Capability{
				capabilities.Capabilities[capabilities.DECRYPT_TALLY_WITH_PROOF_ID],
			},
		},
	}
	ctx := context.Background()
	client, closer := server(ctx)
	defer closer()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			requestUUID, err := uuid.NewRandom()
			if err != nil {
				t.Fatalf("fail to compute uuid for request: %v", err)
			}
			in := &cs.ListCapabilitiesRequest{
				RequestId:  requestUUID.String(),
				FilterName: tc.filterName,
				FilterId:   tc.filterId,
				FilterArgs: tc.filterArgs,
				PageNum:    tc.pageNum,
				PageSize:   tc.pageSize,
			}

			output, err := client.ListCapabilities(ctx, in)
			if err != nil {
				t.Fatalf("client.ListCapabilities failed: %v", err)
			}

			if output.GetRequestId() != requestUUID.String() {
				t.Fatalf("wrong RequestId: expected \"%s\", got \"%s\"", requestUUID, output.GetRequestId())
			}
			if !output.Ok {
				t.Fatalf("expected Ok to be true")
			}
			if output.GetError() != "" {
				t.Fatalf("expected empty Error")
			}
			if len(output.GetCapabilities()) != len(tc.expectedCapabilities) {
				t.Fatalf("wrong number of  Capabilities: expected %d, got %d", len(tc.expectedCapabilities), len(output.Capabilities))
			}
			for i, capability := range tc.expectedCapabilities {
				if capability.Id != output.Capabilities[i].Id {
					t.Fatalf("expected Id \"%s\", got \"%s\"", capability.Id, output.Capabilities[i].Id)
				}
				if capability.Description != output.Capabilities[i].Description {
					t.Fatalf("expected Description \"%s\", got \"%s\"", capability.Description, output.Capabilities[i].Description)
				}
				if capability.Name != output.Capabilities[i].Name {
					t.Fatalf("expected Name \"%s\", got \"%s\"", capability.Name, output.Capabilities[i].Name)
				}
				if capability.NumArguments != output.Capabilities[i].NumArguments {
					t.Fatalf("expected NumArguments \"%d\", got \"%d\"", capability.NumArguments, output.Capabilities[i].NumArguments)
				}
			}
		})
	}
}
