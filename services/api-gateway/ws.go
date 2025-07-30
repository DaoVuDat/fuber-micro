package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/proto/driver"
)

var (
	connManager = messaging.NewConnectionManager()
)

func handlerDriversWebSocket(rb *messaging.RabbitMQ) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := connManager.Upgrade(w, r)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		userID := r.URL.Query().Get("userID")
		if userID == "" {
			log.Printf("UserID is required")
			return
		}

		packageSlug := r.URL.Query().Get("packageSlug")
		if userID == "" {
			log.Printf("No packageSlug provided")
			return
		}

		// Add connection to manager
		connManager.Add(userID, conn)

		ctx := r.Context()

		driverService, err := grpc_clients.NewDriverServiceClient()
		if err != nil {
			log.Fatal(err)
		}

		// Closing connections
		defer func() {

			connManager.Remove(userID)

			driverService.Client.UnRegisterDriver(ctx, &driver.RegisterDriverRequest{
				DriverID:    userID,
				PackageSlug: packageSlug,
			})
			driverService.Close()
			log.Println("Driver unregistered: ", userID)
		}()

		driverData, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})

		if err != nil {
			log.Printf("Error registering driver: %v", err)
			return
		}

		if err := connManager.SendMessage(userID, contracts.WSMessage{
			Type: contracts.DriverCmdRegister,
			Data: driverData.Driver,
		}); err != nil {
			log.Printf("Error sending message: %v", err)
			return
		}

		// Initialize the queue consumers
		queues := []string{
			messaging.DriverCmdTripRequestQueue,
		}

		// Send message from RabbitMQ to websocket (to frontend)
		for _, queue := range queues {
			consumer := messaging.NewQueueConsumer(rb, connManager, queue)

			if err := consumer.Start(); err != nil {
				log.Printf("Failed to start consumer for queue %s: %v", queue, err)
				return
			}
		}

		// Read Message from frontend
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message: %v", err)
				break
			}

			type DriverMessage struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}

			var driverMsg DriverMessage
			if err := json.Unmarshal(message, &driverMsg); err != nil {
				log.Printf("Failed to unmarshal driver message: %v", err)
				continue
			}

			// handle the different types of messages
			switch driverMsg.Type {
			case contracts.DriverCmdLocation:
				// Handle driver location update in the future
				continue
			case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
				// Forward the message to the rabbitmq
				if err := rb.PublishMessage(ctx, driverMsg.Type, contracts.AmqpMessage{
					OwnerID: userID,
					Data:    driverMsg.Data,
				}); err != nil {
					log.Printf("Failed to publish message to RabbitMQ: %v", err)
					continue
				}
			default:
				log.Printf("Unknown driver message type: %s", driverMsg.Type)
			}
		}
	}
}

func handlerRidersWebSocket(rb *messaging.RabbitMQ) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := connManager.Upgrade(w, r)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		userID := r.URL.Query().Get("userID")
		if userID == "" {
			log.Printf("UserID is required")
			return
		}

		// Add connection to manager
		connManager.Add(userID, conn)
		defer connManager.Remove(userID)

		// Initialize the queue consumers
		queues := []string{
			messaging.NotifyDriverNoDriversFoundQueue,
			messaging.NotifyDriverAssignQueue,
		}

		// Send message from RabbitMQ to websocket (to frontend)
		for _, queue := range queues {
			consumer := messaging.NewQueueConsumer(rb, connManager, queue)

			if err := consumer.Start(); err != nil {
				log.Printf("Failed to start consumer for queue %s: %v", queue, err)
				return
			}
		}

		// Read message from frontend
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				break
			}
			log.Printf("Received message: %s", message)
		}
	}
}
