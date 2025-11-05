package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/forgoes/logging"

	"github.com/forgoes/visa/grpc"
	"github.com/forgoes/visa/http"
	"github.com/forgoes/visa/runtime"
)

var l *logging.Logger

func init() {
	l = logging.GetRootLogger()
}

func remaining(ctx context.Context) time.Duration {
	if d, ok := ctx.Deadline(); ok {
		return time.Until(d)
	}
	return 0
}

func shutdown(rt *runtime.Runtime, grpcServer *grpc.Server, gateway *grpc.Gateway, httpServer *http.Server) {
	l.Info().Logf("[exit] stopping with timeout %v seconds...", rt.Config.App.StopTimeout)

	stopTimeout := time.Duration(rt.Config.App.StopTimeout) * time.Second
	stopCtx, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()

	start := time.Now()

	// stage 1
	// goroutine 1: stop gateway, grpc
	// goroutine 2: stop http
	g, ctx := errgroup.WithContext(stopCtx)
	g.Go(func() error {
		// l.Info().Logf("[exit] stopping gateway (remaining=%v)...", remaining(ctx))

		stageStart := time.Now()
		if err := gateway.Stop(ctx); err != nil {
			return err
		}
		l.Info().Logf("[exit] gateway stopped in %v", time.Since(stageStart))

		if remaining(ctx) <= 0 {
			return context.DeadlineExceeded
		}

		// l.Info().Logf("[exit] stopping grpc (remaining=%v)...", remaining(ctx))

		stageStart = time.Now()
		if err := grpcServer.Stop(ctx); err != nil {
			return err
		}
		l.Info().Logf("[exit] grpc stopped in %v", time.Since(stageStart))

		return nil
	})

	g.Go(func() error {
		// l.Info().Logf("[exit] stopping http (remaining=%v)...", remaining(ctx))

		stageStart := time.Now()
		if err := httpServer.Stop(ctx); err != nil {
			return err
		}
		l.Info().Logf("[exit] http stopped in %v", time.Since(stageStart))

		return nil
	})

	if err := g.Wait(); err != nil {
		l.Error().Logf("[exit] stage 1 stopped error %v", err)
	}
	// l.Info().Logf("[exit] stage 1 done, elapsed=%v, remaining=%v", time.Since(start), remaining(stopCtx))

	stageStart := time.Now()
	newRemain := remaining(stopCtx)
	if newRemain <= 0 {
		l.Error().Logf("[exit] skipped (timeout exceeded)")
		return
	}

	ctxRuntime, cancelRuntime := context.WithTimeout(context.Background(), newRemain)
	defer cancelRuntime()

	if err := rt.Close(ctxRuntime); err != nil {
		l.Error().Logf("[exit] runtime error %v", err)
	}
	l.Info().Logf("[exit] runtime stopped in %v", time.Since(stageStart))

	stageStart = time.Now()
	l.Flush()
	l.Close()
	println(fmt.Sprintf("[exit] logging stopped in %v", time.Since(stageStart)))

	println(fmt.Sprintf("[exit] shutdown successfully, total elapsed %v", time.Since(start)))

	os.Exit(0)
}

func main() {
	rt, err := runtime.NewRuntime()
	if err != nil {
		l.Panic().E(err).Log()
	}

	grpcServer := grpc.NewServer(rt)
	grpcGateway := grpc.NewGateway(rt)
	httpServer := http.NewServer(rt)

	errCh := make(chan error, 3)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	l.Info().Logf("starting grpc:    %s", fmt.Sprintf("%s:%d", rt.Config.Serve.GRPC.Host, rt.Config.Serve.GRPC.Port))
	l.Info().Logf("starting gateway: %s, binding: %s",
		fmt.Sprintf("%s:%d", rt.Config.Serve.Gateway.Host, rt.Config.Serve.Gateway.Port),
		rt.Config.Serve.Gateway.Upstream,
	)
	l.Info().Logf("starting http:    %s", fmt.Sprintf("%s:%d", rt.Config.Serve.HTTP.Host, rt.Config.Serve.HTTP.Port))
	go grpcServer.Serve(errCh)
	go grpcGateway.Serve(errCh)
	go httpServer.Serve(errCh)

	select {
	case sig := <-sigCh:
		l.Info().Logf("[exit] system signal: %s received, will shut down gracefully", sig.String())
		shutdown(rt, grpcServer, grpcGateway, httpServer)
	case err := <-errCh:
		if err != nil {
			l.Error().Logf("[exit] server error: %v received, will shut down gracefully", err)
		} else {
			l.Info().Logf("[exit] server exited, will shut down gracefully")
		}
		shutdown(rt, grpcServer, grpcGateway, httpServer)
	}
}
