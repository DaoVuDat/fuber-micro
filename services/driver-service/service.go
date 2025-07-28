package main

import (
	pb "ride-sharing/shared/proto/driver"
)

type Service struct {
	drivers []*pb.Driver
}

type driverInMap struct {
	Driver *pb.Driver
}

func NewService() *Service {
	return &Service{
		drivers: make([]*pb.Driver, 0),
	}
}

// TODO: implement register and unregister drivers
