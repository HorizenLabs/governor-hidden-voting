package crypto_service

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const Name = "e-voting"
const Version = "0.1.0"

type Capability interface {
	Id() string
	Name() string
	Description() string
	NumArguments() uint32
	Compute([][]byte) ([][]byte, error)
}

func NewCryptographicService() *CryptographicService {
	return &CryptographicService{
		capabilities: make(map[string]Capability),
	}
}

type CryptographicService struct {
	cs.UnimplementedCryptographicServiceServer
	capabilities map[string]Capability
}

func (s *CryptographicService) Watch(req *cs.HealthCheckRequest, stream cs.CryptographicService_WatchServer) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		if err := stream.Send(&cs.HealthCheckResponse{Status: cs.HealthCheckResponse_ACTIVE}); err != nil {
			return err
		}
	}
}

func (s *CryptographicService) ListCapabilities(ctx context.Context, req *cs.ListCapabilitiesRequest) (*cs.ListCapabilitiesResponse, error) {
	capabilities := s.listRawCapabilities()
	capabilities = filterCapabilities(capabilities, req.FilterName, req.FilterId, req.FilterArgs)
	capabilities = paginateCapabilities(capabilities, req.PageSize, req.PageNum)

	return &cs.ListCapabilitiesResponse{
		RequestId:    req.GetRequestId(),
		Ok:           true,
		Capabilities: capabilities,
	}, nil
}

func (s *CryptographicService) Compute(ctx context.Context, exec *cs.Execute) (*cs.Result, error) {
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

func (s *CryptographicService) RegisterCapability(capability Capability) error {
	if _, exists := s.capabilities[capability.Id()]; !exists {
		s.capabilities[capability.Id()] = capability
		return nil
	} else {
		return fmt.Errorf("a capability with id \"%s\" has already been registered", capability.Id())
	}
}

func (s *CryptographicService) Serve(lis net.Listener) error {
	grpcServer := grpc.NewServer()
	cs.RegisterCryptographicServiceServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

func (s *CryptographicService) RegisterToProvingGrounds(
	lis net.Listener,
	protocol string,
	register_address string,
	register_port int,
) error {
	serverAddr := fmt.Sprintf("%v:%v", register_address, register_port)
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(serverAddr, dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := cs.NewCryptographicRegistrarClient(conn)
	primitiveUUID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to compute uuid for primitive: %v", err)
	}
	capabilitiesResponse, err := s.ListCapabilities(context.Background(), &cs.ListCapabilitiesRequest{})
	if err != nil {
		return err
	}
	if !capabilitiesResponse.GetOk() {
		return fmt.Errorf(capabilitiesResponse.GetError())
	}

	addressAndPort := strings.Split(lis.Addr().String(), ":")
	address := addressAndPort[0]
	port, err := strconv.Atoi(addressAndPort[1])
	if err != nil {
		return err
	}
	primitiveInfo := cs.RegisterCryptographyPrimitive{
		RequestId:    primitiveUUID.String(),
		Version:      Version,
		Name:         Name,
		Address:      fmt.Sprintf("%v://%v", protocol, address),
		Port:         uint32(port),
		Capabilities: capabilitiesResponse.Capabilities,
	}

	registrationResult, err := client.Register(context.Background(), &primitiveInfo)
	if err != nil {
		return fmt.Errorf("failed to register primitive: %v", err)
	}
	if !registrationResult.GetOk() {
		return fmt.Errorf(registrationResult.GetError())
	}
	return nil
}

func (s *CryptographicService) listRawCapabilities() []*cs.Capability {
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
