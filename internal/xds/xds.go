package xds

import (
	"context"
	"fmt"
	"net"

	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
	xdsgrpc "github.com/110y/bootes/internal/xds/internal/grpc"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     logr.Logger
}

func NewServer(ctx context.Context, sc xdscache.SnapshotCache, c cache.Cache, s store.Store, l logr.Logger, config *Config) (*Server, error) {
	srv := server.NewServer(ctx, sc, newCallbacks(c, s, l.WithName("callbacks")))

	gc := &xdsgrpc.Config{
		EnableChannelz:   config.EnableGRPCChannelz,
		EnableReflection: config.EnableGRPCReflection,
	}
	gs := xdsgrpc.NewServer(ctx, srv, gc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d :%w", config.Port, err)
	}

	return &Server{
		grpcServer: gs,
		listener:   lis,
		logger:     l,
	}, nil
}

func NewSnapshotCache(l logr.Logger) xdscache.SnapshotCache {
	return xdscache.NewSnapshotCache(true, xdscache.IDHash{}, newSnapshotCacheLogger(l))
}

func (s *Server) Start(stopCh chan struct{}) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("starting xds server")
		errCh <- s.grpcServer.Serve(s.listener)
	}()

	select {
	case <-stopCh:
		s.logger.Info("stopping xds server")
		s.grpcServer.Stop()
		return nil
	case err := <-errCh:
		return err
	}
}
