package server

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/harche/crio-mcp-server/pkg/cgroup"
	"github.com/harche/crio-mcp-server/pkg/cri"
	"github.com/harche/crio-mcp-server/pkg/journal"
	pb "github.com/harche/crio-mcp-server/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MCPServer struct {
	pb.UnimplementedMCPServiceServer
	ConfigPath string
	runtime    *cri.Crictl
}

func New(configPath string) *MCPServer {
	return &MCPServer{ConfigPath: configPath, runtime: cri.New("")}
}

func (s *MCPServer) GetCrioConfig(ctx context.Context, _ *pb.Empty) (*pb.CrioConfigResponse, error) {
	data, err := os.ReadFile(s.ConfigPath)
	if err != nil {
		log.Printf("error reading config: %v", err)
		return nil, status.Errorf(codes.Internal, "unable to read config")
	}
	return &pb.CrioConfigResponse{Config: string(data)}, nil
}

func (s *MCPServer) GetRuntimeStatus(ctx context.Context, _ *pb.Empty) (*pb.RuntimeStatusResponse, error) {
	out, err := s.runtime.RuntimeStatus()
	if err != nil {
		log.Printf("runtime status error: %v", err)
		return nil, status.Errorf(codes.Internal, "runtime status error")
	}
	return &pb.RuntimeStatusResponse{Status: out}, nil
}

func (s *MCPServer) ListContainers(ctx context.Context, _ *pb.Empty) (*pb.ContainersResponse, error) {
	out, err := s.runtime.ListContainers()
	if err != nil {
		log.Printf("list containers error: %v", err)
		return nil, status.Errorf(codes.Internal, "list containers error")
	}
	return &pb.ContainersResponse{Containers: out}, nil
}

func (s *MCPServer) InspectContainer(ctx context.Context, req *pb.ContainerRequest) (*pb.ContainerInspectResponse, error) {
	out, err := s.runtime.InspectContainer(req.GetId())
	if err != nil {
		log.Printf("inspect container error: %v", err)
		return nil, status.Errorf(codes.Internal, "inspect container error")
	}
	return &pb.ContainerInspectResponse{Info: out}, nil
}

func (s *MCPServer) GetContainerStats(ctx context.Context, req *pb.ContainerRequest) (*pb.ContainerStatsResponse, error) {
	info, err := s.runtime.InspectContainer(req.GetId())
	if err != nil {
		log.Printf("inspect container error: %v", err)
		return nil, status.Errorf(codes.Internal, "inspect container error")
	}

	stats, err := cgroup.StatsFromInspect(info)
	if err != nil {
		log.Printf("stats error: %v", err)
		return nil, status.Errorf(codes.Internal, "stats error")
	}

	return &pb.ContainerStatsResponse{
		CpuUsageUsec:     stats.CPUUsageUSec,
		MemoryUsageBytes: stats.MemoryUsageBytes,
	}, nil
}

func (s *MCPServer) GetContainerConfig(ctx context.Context, req *pb.ContainerRequest) (*pb.ContainerConfigResponse, error) {
	cfg, err := cri.ReadContainerConfig(req.GetId())
	if err != nil {
		log.Printf("read container config error: %v", err)
		return nil, status.Errorf(codes.Internal, "container config error")
	}
	return &pb.ContainerConfigResponse{Config: cfg}, nil
}

func parseTime(val string) (time.Time, error) {
	if val == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (s *MCPServer) GetLogs(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	since, err := parseTime(req.GetSince())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid since")
	}
	until, err := parseTime(req.GetUntil())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid until")
	}
	logs, err := journal.ReadLogs(journal.Options{
		Unit:        req.GetUnit(),
		ContainerID: req.GetContainerId(),
		Since:       since,
		Until:       until,
		Priority:    req.GetPriority(),
		Limit:       int(req.GetTail()),
	})
	if err != nil {
		log.Printf("log fetch error: %v", err)
		return nil, status.Errorf(codes.Internal, "log fetch error")
	}
	return &pb.LogResponse{Logs: logs}, nil
}

func (s *MCPServer) StreamLogs(req *pb.LogRequest, stream pb.MCPService_StreamLogsServer) error {
	since, err := parseTime(req.GetSince())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid since")
	}
	until, err := parseTime(req.GetUntil())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid until")
	}
	opts := journal.Options{
		Unit:        req.GetUnit(),
		ContainerID: req.GetContainerId(),
		Since:       since,
		Until:       until,
		Priority:    req.GetPriority(),
		Follow:      true,
	}
	return journal.StreamLogs(opts, func(line string) error {
		return stream.Send(&pb.LogEntry{Line: line})
	})
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
