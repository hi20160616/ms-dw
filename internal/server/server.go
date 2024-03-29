package server

import (
	"context"
	"log"
	"net"

	pb "github.com/hi20160616/fetchnews-api/proto/v1"
	"github.com/hi20160616/ms-dw/configs"
	"github.com/hi20160616/ms-dw/internal/service"
	"google.golang.org/grpc"
)

var s = grpc.NewServer()

func Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", configs.Data.MS["dw"].Addr)
	if err != nil {
		return err
	}
	pb.RegisterFetchNewsServer(s, &service.Server{})
	return s.Serve(lis)
}

func Stop(ctx context.Context) error {
	s.GracefulStop()
	log.Printf("grpc server gracefully stopped.")
	return nil
}
