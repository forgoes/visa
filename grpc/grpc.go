package grpc

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/net/netutil"
	"google.golang.org/grpc"

	v1 "github.com/forgoes/visa/grpc/v1"
	"github.com/forgoes/visa/runtime"
)

type Server struct {
	rt     *runtime.Runtime
	server *grpc.Server
}

func NewServer(rt *runtime.Runtime) *Server {
	server := grpc.NewServer(grpc.MaxConcurrentStreams(rt.Config.Serve.GRPC.MaxConcurrentStreams))

	v1.RegisterService(rt, server)

	s := &Server{
		rt:     rt,
		server: server,
	}

	return s
}

func (s *Server) Serve(errCh chan<- error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.rt.Config.Serve.GRPC.Host, s.rt.Config.Serve.GRPC.Port))
	if err != nil {
		errCh <- err
		return
	}

	lis = netutil.LimitListener(lis, s.rt.Config.Serve.GRPC.MaxTcpConnections)

	errCh <- s.server.Serve(lis)
}

func (s *Server) Stop(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}
