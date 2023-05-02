package main

import (
	"context"
	"fmt"
	"time"

	cpb "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/capabilities"
	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
)

func newServer() *cryptographicServiceServer {
	return &cryptographicServiceServer{}
}

type cryptographicServiceServer struct {
	cs.UnimplementedCryptographicServiceServer
}

func (s *cryptographicServiceServer) Watch(req *cs.HealthCheckRequest, stream cs.CryptographicService_WatchServer) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		if err := stream.Send(&cs.HealthCheckResponse{Status: cs.HealthCheckResponse_ACTIVE}); err != nil {
			return err
		}
	}
}

func (s *cryptographicServiceServer) ListCapabilities(ctx context.Context, req *cs.ListCapabilitiesRequest) (*cs.ListCapabilitiesResponse, error) {
	capabilities := []*cs.Capability{}
	for _, capability := range cpb.ListCapabilities() {
		result := true
		if filterName := req.FilterName; filterName != nil {
			result = result && (*filterName == capability.Name)
		}
		if filterId := req.FilterId; filterId != nil {
			result = result && (*filterId == capability.Id)
		}
		if filterArgs := req.FilterArgs; filterArgs != nil {
			result = result && (*filterArgs == capability.NumArguments)
		}
		if result {
			capabilities = append(capabilities, capability)
		}
	}

	pageSize := req.PageSize
	if pageSize == nil {
		pageSize = new(uint32)
		*pageSize = 50
	}
	if pageNum := req.PageNum; pageNum != nil {
		startIndex := (*pageNum * *pageSize)
		endIndex := startIndex + *pageSize
		capabilitiesLen := uint32(len(capabilities))
		if endIndex > capabilitiesLen {
			endIndex = capabilitiesLen
		}
		if startIndex < capabilitiesLen {
			capabilities = capabilities[startIndex:endIndex]
		} else {
			capabilities = nil
		}
	}

	return &cs.ListCapabilitiesResponse{
		RequestId:    req.GetRequestId(),
		Ok:           true,
		Capabilities: capabilities,
	}, nil
}

func (s *cryptographicServiceServer) Compute(ctx context.Context, exec *cs.Execute) (*cs.Result, error) {
	ret := new(cs.Result)
	var err error
	var res [][]byte

	ret.RequestId = exec.GetRequestId()

	handler, ok := cpb.Handlers[exec.GetCapabilityId()]
	if ok {
		res, err = handler(exec.GetArguments())
	} else {
		err = fmt.Errorf("requested capability is unsupported")
	}

	if err != nil {
		ret.Ok = false
		errString := err.Error()
		ret.Error = &errString
	} else {
		ret.Ok = true
		ret.Result = make([][]byte, len(res))
		copy(ret.Result, res)
	}
	return ret, nil
}
