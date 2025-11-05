package v1

import (
	"context"

	pb "github.com/forgoes/proto-go/api/visa/v1"
	"google.golang.org/grpc"

	"github.com/forgoes/visa/runtime"
)

type Service struct {
	rt *runtime.Runtime

	pb.UnimplementedVisaServiceServer
}

func RegisterService(rt *runtime.Runtime, server *grpc.Server) {
	service := &Service{
		rt: rt,
	}

	pb.RegisterVisaServiceServer(server, service)
}

func (s *Service) Echo(_ context.Context, in *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{Value: in.Value + " world"}, nil
}
