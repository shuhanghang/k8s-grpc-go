package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/shuhanghang/k8s-grpc-go/pb"
	"github.com/shuhanghang/k8s-grpc-go/utils"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 22222, "The port of grpc server listening")
)

type server struct {
	pb.UnimplementedExampleServiceServer
}

// service implement
func (s *server) Service(ctx context.Context, req *pb.ExampleRequest) (*pb.ExampleResponse, error) {
	hostName, _ := os.Hostname()
	ip := utils.GetIp()
	result := fmt.Sprintf("hostName: %s, ip: %s", hostName, ip)
	return &pb.ExampleResponse{Result: result}, nil
}

func main() {
	flag.Parse()
	listAddr := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", listAddr)
	if err != nil {
		log.Fatalf("failed to listen: %s", err.Error())
	}

	s := grpc.NewServer()

	pb.RegisterExampleServiceServer(s, &server{})

	fmt.Printf("Starting gprc server: %s\n", listAddr)
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to run gprc service: %s", err.Error())
	}
}
