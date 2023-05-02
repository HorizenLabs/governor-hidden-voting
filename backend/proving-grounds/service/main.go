package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	cpb "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/capabilities"
	cs "github.com/HorizenLabs/e-voting-poc/backend/proving-grounds/crypto_server"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const Name = "e-voting"
const Version = "0.1.0"

var (
	protocol         = flag.String("protocol", "http", "The protocol to use")
	address          = flag.String("address", "127.0.0.1", "The reachable address")
	bind_address     = flag.String("bind_address", "127.0.0.1", "The bind address")
	port             = flag.Uint("port", 3333, "The bind port")
	register_address = flag.String("register_address", "127.0.0.1", "The address to proving grounds registration")
	register_port    = flag.Uint("register_port", 5678, "The port to proving grounds registration")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *bind_address, *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	var serverOpts []grpc.ServerOption

	serverAddr := fmt.Sprintf("%v:%v", *register_address, *register_port)
	conn, err := grpc.Dial(serverAddr, dialOpts...)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	client := cs.NewCryptographicRegistrarClient(conn)

	primitiveUUID, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("failed to compute uuid for primitive: %v", err)
	}
	primitiveInfo := cs.RegisterCryptographyPrimitive{
		RequestId:    primitiveUUID.String(),
		Version:      Version,
		Name:         Name,
		Address:      fmt.Sprintf("%v://%v", *protocol, *address),
		Port:         uint32(*port),
		Capabilities: cpb.ListCapabilities(),
	}

	_, err = client.Register(context.Background(), &primitiveInfo)
	if err != nil {
		log.Fatalf("failed to register primitive: %v", err)
	}

	grpcServer := grpc.NewServer(serverOpts...)
	cs.RegisterCryptographicServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
