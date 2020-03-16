//
// This file is generated by Arion.
//

package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

var (
	grpcPort = flag.String("grpcport", ":50051", "gRPC server port")
	httpPort = flag.String("httpport", ":50052", "HTTP server port")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	{{ .PBRegisterServerFunc -}}(s, &server{})

	// start sidecar http server
	go func() {
		http.HandleFunc("/resp", handleUpdateResponse)
		log.Fatalf("failed http server: %v", http.ListenAndServe(*httpPort, nil))
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
