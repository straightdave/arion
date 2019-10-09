package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "github.com/straightdave/arion/test_server/helloworld"
)

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

const (
	port = ":50051"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: name=%v, gender=%v", in.GetName(), in.GetGender())
	return &pb.HelloReply{Message: "Hello " + in.GetName(), Gender: in.GetGender()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("Starting server at port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
