package server

import (
	"context"

	"google.golang.org/grpc"
)

// Registrar is interface registers self into grpcServer
type Registrar interface {
	// Register registers self into grpcServer
	Register(ctx context.Context, grpcServer *grpc.Server) (err error)
}
