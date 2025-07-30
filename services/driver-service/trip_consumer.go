package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type tripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ, service *Service) *tripConsumer {
	return &tripConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *tripConsumer) Listen(ctx context.Context) error {
	return c.rabbitmq.ConsumeMessages(messaging.FindAvailableDriverQueue, func(ctx context.Context, msg amqp.Delivery) error {
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return nil
		}

		var payload messaging.TripEventData
		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return nil
		}

		log.Printf("driver service received a message: %+v", payload)

		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return c.handleFindAndNotifyDrivers(ctx, payload)
		}

		log.Printf("unknown trip event key: %+v", payload)

		return nil
	})
}

func (c *tripConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload messaging.TripEventData) error {
	log.Println("[handleFindAndNotifyDrivers] Get length:", c.service.GetLength())
	log.Println("[handleFindAndNotifyDrivers]", payload)
	suitableDrivers := c.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)
	log.Printf("Found suitable drivers: %d", len(suitableDrivers))

	if len(suitableDrivers) == 0 {
		// Notify the driver that no drivers are available
		if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: payload.Trip.UserID,
		}); err != nil {
			log.Printf("failed to publish no drivers found message: %v", err)
			return err
		}

		return nil
	}

	suitableDriversId := suitableDrivers[0]

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// notify the driver about potential trip
	if err := c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: suitableDriversId,
		Data:    marshalledEvent,
	}); err != nil {
		log.Printf("failed to publish trip request message: %v", err)
		return err
	}

	return nil
}
