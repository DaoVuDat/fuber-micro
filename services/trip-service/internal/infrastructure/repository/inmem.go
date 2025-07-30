package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context, fare *domain.RideFareModel) error {
	r.rideFares[fare.ID.Hex()] = fare
	return nil
}

func (r *inmemRepository) GetRiderFareByID(ctx context.Context, fareID string) (*domain.RideFareModel, error) {
	rideFare, ok := r.rideFares[fareID]
	if !ok {
		return nil, fmt.Errorf("ride fare with id %s not found", fareID)
	}
	return rideFare, nil
}

func (r *inmemRepository) GetTripByID(ctx context.Context, tripID string) (*domain.TripModel, error) {
	trip, ok := r.trips[tripID]
	if !ok {
		return nil, fmt.Errorf("trip with id %s not found", tripID)
	}
	return trip, nil
}

func (r *inmemRepository) UpdateTrip(ctx context.Context, tripID, status string, driver *pbd.Driver) error {
	trip, ok := r.trips[tripID]
	if !ok {
		return fmt.Errorf("trip with id %s not found", tripID)
	}

	trip.Status = status
	if trip.Driver != nil {
		trip.Driver = &pb.TripDriver{
			Id:             driver.Id,
			Name:           driver.Name,
			CarPlate:       driver.CarPlate,
			ProfilePicture: driver.ProfilePicture,
		}
	}

	return nil
}
