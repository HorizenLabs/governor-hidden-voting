package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/capabilities"
	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const Name = "e-voting"
const Version = "0.1.0"

type Args struct {
	protocol         string
	address          string
	bind_address     string
	port             uint
	register_address string
	register_port    uint
}

func newDefaultArgs() *Args {
	return &Args{
		protocol:         "http",
		address:          "127.0.0.1",
		bind_address:     "127.0.0.1",
		port:             3333,
		register_address: "127.0.0.1",
		register_port:    5678,
	}
}

func parseArgs() *Args {
	args := new(Args)
	defaultArgs := newDefaultArgs()

	flag.StringVar(&args.protocol, "protocol", defaultArgs.protocol, "The protocol to use")
	flag.StringVar(&args.address, "address", defaultArgs.address, "The reachable address")
	flag.StringVar(&args.bind_address, "bind_address", defaultArgs.bind_address, "The bind address")
	flag.UintVar(&args.port, "port", defaultArgs.port, "The bind port")
	flag.StringVar(&args.register_address, "register_address", defaultArgs.register_address, "The address to proving grounds registration")
	flag.UintVar(&args.register_port, "register_port", defaultArgs.register_port, "The port to proving grounds registration")

	flag.Parse()
	return args
}

func registerAndServeEVotingService(args *Args) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", args.bind_address, args.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	serverAddr := fmt.Sprintf("%v:%v", args.register_address, args.register_port)
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

	service := newCryptographicServiceServer()
	capabilities := []Capability{
		capabilities.NewKeyPairWithProofCapability{},
		capabilities.EncryptVoteWithProofCapability{},
		capabilities.DecryptTallyWithProofCapability{},
	}
	for _, capability := range capabilities {
		service.registerCapability(capability)
	}

	primitiveInfo := cs.RegisterCryptographyPrimitive{
		RequestId:    primitiveUUID.String(),
		Version:      Version,
		Name:         Name,
		Address:      fmt.Sprintf("%v://%v", args.protocol, args.address),
		Port:         uint32(args.port),
		Capabilities: service.listRawCapabilities(),
	}

	registrationResult, err := client.Register(context.Background(), &primitiveInfo)
	if err != nil {
		return fmt.Errorf("failed to register primitive: %v", err)
	}
	if !registrationResult.GetOk() {
		return fmt.Errorf(registrationResult.GetError())
	}

	grpcServer := grpc.NewServer()
	cs.RegisterCryptographicServiceServer(grpcServer, newCryptographicServiceServer())
	grpcServer.Serve(lis)
	return nil
}

func main() {
	args := parseArgs()
	if err := registerAndServeEVotingService(args); err != nil {
		log.Fatal(err)
	}
}
