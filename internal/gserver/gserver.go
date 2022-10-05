// Package gserver used for gRPC server side communications
package gserver

import (
	"context"
	"log"
	"net"

	"github.com/zklevsha/go-musthave-devops/internal/pb"
	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedMonitoringServer
	Storage structs.Storage
}

func (s *server) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	log.Printf("INFO Received gRPC request: %v, params: %s", in.ProtoReflect().Descriptor().FullName(), in.String())
	m, err := serializer.DecodeGRPCMetric(in.Metric)
	if err != nil {
		response := pb.Response{Message: "", Error: err.Error()}
		return &pb.UpdateMetricResponse{Response: &response}, nil
	}

	err = s.Storage.UpdateMetric(m)
	if err != nil {
		response := pb.Response{Message: "", Error: err.Error()}
		return &pb.UpdateMetricResponse{Response: &response}, nil
	}

	response := pb.Response{Message: "Metric was updated", Error: ""}
	return &pb.UpdateMetricResponse{Response: &response}, nil
}

func Start(socket string, store structs.Storage) {
	listener, err := net.Listen("tcp", socket)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterMonitoringServer(s, &server{Storage: store})
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
