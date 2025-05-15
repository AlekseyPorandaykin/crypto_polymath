package grpc

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/grpc/action"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	s    *grpc.Server
	port uint
}

func NewServer(port uint, actionHandler *ActionHandler) (*Server, error) {
	grpcServer := grpc.NewServer()
	action.RegisterActionServiceServer(grpcServer, actionHandler)
	return &Server{s: grpcServer, port: port}, nil
}

func (s *Server) Run(_ context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	return s.s.Serve(listener)
}
