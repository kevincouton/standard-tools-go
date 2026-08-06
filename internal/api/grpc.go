package api

import (
	"context"

	pb "github.com/kevincouton/standard-tools-go/proto/health"
)

type HealthServer struct {
	pb.UnimplementedHealthServer
}

func (s *HealthServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: "ok"}, nil
}
