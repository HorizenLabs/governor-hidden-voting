package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/capabilities"
	"github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_service"
)

type Args struct {
	protocol         string
	address          string
	port             uint
	register_address string
	register_port    uint
}

func newDefaultArgs() *Args {
	return &Args{
		protocol:         "http",
		address:          "127.0.0.1",
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
	flag.UintVar(&args.port, "port", defaultArgs.port, "The bind port")
	flag.StringVar(&args.register_address, "register_address", defaultArgs.register_address, "The address to proving grounds registration")
	flag.UintVar(&args.register_port, "register_port", defaultArgs.register_port, "The port to proving grounds registration")

	flag.Parse()
	return args
}

func newEvotingService() *crypto_service.CryptographicService {
	service := crypto_service.NewCryptographicService()
	capabilities := []crypto_service.Capability{
		capabilities.NewKeyPairWithProofCapability{},
		capabilities.EncryptVoteWithProofCapability{},
		capabilities.DecryptTallyWithProofCapability{},
	}
	for _, capability := range capabilities {
		service.RegisterCapability(capability)
	}
	return service
}

func main() {
	args := parseArgs()
	service := newEvotingService()
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", args.address, args.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	err = service.RegisterToProvingGrounds(lis, args.protocol, args.register_address, int(args.register_port))
	if err != nil {
		log.Fatal(err)
	}
	err = service.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
