package grpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/forgoes/proto-go/api/visa/v1"

	"github.com/forgoes/visa/runtime"
)

type Gateway struct {
	rt     *runtime.Runtime
	mux    *gw.ServeMux
	server *http.Server
}

func NewGateway(rt *runtime.Runtime) *Gateway {
	mux := gw.NewServeMux()

	return &Gateway{
		rt:  rt,
		mux: mux,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", rt.Config.Serve.Gateway.Host, rt.Config.Serve.Gateway.Port),
			Handler: mux,
		},
	}
}

func (g *Gateway) Serve(errCh chan<- error) {
	conn, err := grpc.NewClient(
		g.rt.Config.Serve.Gateway.Upstream,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		errCh <- err
		return
	}

	err = pb.RegisterVisaServiceHandler(context.Background(), g.mux, conn)
	if err != nil {
		errCh <- err
		return
	}

	if err := g.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errCh <- err
	}
}

func (g *Gateway) Stop(ctx context.Context) (err error) {
	if err := g.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
