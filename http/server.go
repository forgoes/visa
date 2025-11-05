package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/forgoes/visa/runtime"
)

type Server struct {
	rt     *runtime.Runtime
	server *http.Server
}

func NewServer(rt *runtime.Runtime) *Server {
	if rt.Config.Serve.HTTP.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, nil)
	})

	addr := fmt.Sprintf("%s:%d", rt.Config.Serve.HTTP.Host, rt.Config.Serve.HTTP.Port)

	return &Server{
		rt: rt,
		server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func (s *Server) Serve(errCh chan<- error) {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errCh <- err
	}
}

func (s *Server) Stop(ctx context.Context) (err error) {
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
