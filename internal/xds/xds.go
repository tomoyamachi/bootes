package xds

import (
	"context"
	"fmt"
	"net"

	"github.com/110y/bootes/internal/cache"
	"github.com/110y/bootes/internal/k8s/store"
	xdsgrpc "github.com/110y/bootes/internal/xds/internal/grpc"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer    *grpc.Server
	listener      net.Listener
	snapshotCache *cache.Cache
	k8sStore      store.Store
}

func NewServer(ctx context.Context, snapshotCache *cache.Cache, k8sStore store.Store, config *Config) (*Server, error) {
	gc := &xdsgrpc.Config{
		EnableChannelz:   config.EnableGRPCChannelz,
		EnableReflection: config.EnableGRPCReflection,
	}
	gs := xdsgrpc.NewServer(ctx, snapshotCache, k8sStore, gc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		// TODO: wrap
		return nil, err
	}

	return &Server{
		grpcServer:    gs,
		listener:      lis,
		snapshotCache: snapshotCache,
		k8sStore:      k8sStore,
	}, nil
}

func (s *Server) Start() error {
	// TODO: wrap error
	return s.grpcServer.Serve(s.listener)
}
