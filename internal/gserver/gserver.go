// Package gserver used for gRPC server side communications
package gserver

import "github.com/zklevsha/go-musthave-devops/internal/pb"

type server struct {
	pb.UnimplementedMonitoringServer
}
