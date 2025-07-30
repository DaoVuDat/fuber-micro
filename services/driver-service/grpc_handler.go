package main

import (
	"context"
	"log"
	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedDriverServiceServer
	service *Service
}

func NewGRPCHandler(server *grpc.Server, service *Service) {
	pb.RegisterDriverServiceServer(server, &gRPCHandler{
		service: service,
	})
}

func (h *gRPCHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driver, err := h.service.RegisterDriver(req.GetDriverID(), req.GetPackageSlug())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register driver")
	}

	log.Println("[GRPC RegisterDriver] Driver registered successfully:", driver.Id)
	log.Println("[GRPC RegisterDriver] Length of drivers:", h.service.GetLength())
	return &pb.RegisterDriverResponse{
		Driver: driver,
	}, nil
}

func (h *gRPCHandler) UnregisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	h.service.UnregisterDriver(req.GetDriverID())

	return &pb.RegisterDriverResponse{
		Driver: &pb.Driver{
			Id: req.GetDriverID(),
		},
	}, nil
}
