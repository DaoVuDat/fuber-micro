package domain

import (
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	pb "ride-sharing/shared/proto/trip"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string // ex: van, luxury, sedan
	TotalPriceInCents float64
	Route             *tripTypes.OsrmApiResponse
}

func (f *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                f.ID.Hex(),
		UserID:            f.UserID,
		PackageSlug:       f.PackageSlug,
		TotalPriceInCents: f.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	faresProto := make([]*pb.RideFare, len(fares))
	for i, f := range fares {
		faresProto[i] = f.ToProto()
	}

	return faresProto
}
