package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitMqUri = env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")
)

func main() {
	// RabbitMQ connection
	rb, err := messaging.NewRabbitMQ(rabbitMqUri)
	if err != nil {
		log.Fatal(err)
	}
	defer rb.Close()

	log.Println("Starting API Gateway")
	mux := http.NewServeMux()

	mux.HandleFunc("POST /trip/preview", enableCors(handleTripPreview))
	mux.HandleFunc("POST /trip/start", enableCors(handleTripStart))
	mux.HandleFunc("/ws/drivers", handlerDriversWebSocket(rb))
	mux.HandleFunc("/ws/riders", handlerRidersWebSocket(rb))

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	errorServers := make(chan error)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			errorServers <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errorServers:
		log.Printf("Error starting server: %v", err)
	case sig := <-shutdown:
		log.Printf("Shutting down HTTP server with %v...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down HTTP server: %v", err)
			server.Close()
		}
	}
}
