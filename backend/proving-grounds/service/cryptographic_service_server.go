package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
	"golang.org/x/exp/maps"
)

type Capability interface {
	Id() string
	Name() string
	Description() string
	NumArguments() uint32
	Compute([][]byte) ([][]byte, error)
}

func newCryptographicServiceServer() *cryptographicServiceServer {
	return &cryptographicServiceServer{
		capabilities: make(map[string]Capability),
	}
}

type cryptographicServiceServer struct {
	cs.UnimplementedCryptographicServiceServer
	capabilities map[string]Capability
}

func (s *cryptographicServiceServer) registerCapability(capability Capability) error {
	if _, exists := s.capabilities[capability.Id()]; !exists {
		s.capabilities[capability.Id()] = capability
		return nil
	} else {
		return fmt.Errorf("a capability with id \"%s\" has already been registered", capability.Id())
	}

}

func (s *cryptographicServiceServer) listRawCapabilities() []*cs.Capability {
	keys := maps.Keys(s.capabilities)
	sort.Strings(keys)
	var capabilities []*cs.Capability
	for _, k := range keys {
		capability := &cs.Capability{
			Id:           s.capabilities[k].Id(),
			Name:         s.capabilities[k].Name(),
			Description:  s.capabilities[k].Description(),
			NumArguments: s.capabilities[k].NumArguments(),
		}
		capabilities = append(capabilities, capability)
	}
	return capabilities
}

func filterCapabilities(capIn []*cs.Capability, filterName *string, filterId *string, filterArgs *uint32) []*cs.Capability {
	capOut := []*cs.Capability{}
	for _, cap := range capIn {
		result := true
		if filterName := filterName; filterName != nil {
			result = result && (*filterName == cap.Name)
		}
		if filterId := filterId; filterId != nil {
			result = result && (*filterId == cap.Id)
		}
		if filterArgs := filterArgs; filterArgs != nil {
			result = result && (*filterArgs == cap.NumArguments)
		}
		if result {
			capOut = append(capOut, cap)
		}
	}
	return capOut
}

func paginateCapabilities(capIn []*cs.Capability, pageSize *uint32, pageNum *uint32) []*cs.Capability {
	capOut := capIn[:]
	if pageSize == nil {
		pageSize = new(uint32)
		*pageSize = 50
	}
	if pageNum != nil {
		startIndex := (*pageNum * *pageSize)
		endIndex := startIndex + *pageSize
		capabilitiesLen := uint32(len(capIn))
		if endIndex > capabilitiesLen {
			endIndex = capabilitiesLen
		}
		if startIndex < capabilitiesLen {
			return capIn[startIndex:endIndex]
		} else {
			return []*cs.Capability{}
		}
	}
	return capOut
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
	capabilities := s.listRawCapabilities()
	capabilities = filterCapabilities(capabilities, req.FilterName, req.FilterId, req.FilterArgs)
	capabilities = paginateCapabilities(capabilities, req.PageSize, req.PageNum)

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
	capability := s.capabilities[exec.GetCapabilityId()]
	if capability != nil {
		res, err = capability.Compute(exec.GetArguments())
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
