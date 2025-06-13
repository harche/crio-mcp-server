package server

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/harche/crio-mcp-server/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MCPServer struct {
	pb.UnimplementedMCPServiceServer
	ConfigPath string
}

func New(configPath string) *MCPServer {
	return &MCPServer{ConfigPath: configPath}
}

func (s *MCPServer) GetCrioConfig(ctx context.Context, _ *pb.Empty) (*pb.CrioConfigResponse, error) {
	data, err := os.ReadFile(s.ConfigPath)
	if err != nil {
		log.Printf("error reading config: %v", err)
		return nil, status.Errorf(codes.Internal, "unable to read config")
	}
	return &pb.CrioConfigResponse{Config: string(data)}, nil
}

func (s *MCPServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMCPServiceServer(grpcServer, s)
	log.Printf("gRPC server listening on %s", addr)
	return grpcServer.Serve(lis)
}
