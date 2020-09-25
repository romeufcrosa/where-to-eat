package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	domain "github.com/romeufcrosa/where-to-eat/domain/entities"
	"github.com/romeufcrosa/where-to-eat/providers"
	"github.com/romeufcrosa/where-to-eat/services/api"

	"github.com/bugsnag/bugsnag-go"
)

var (
	listenAddr = os.Getenv("PORT")
)

func main() {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey:          "AIzaSyBvE7lMDfzA9hMypDNfIhGi5VtRbk8HgcU",
		ProjectPackages: []string{"main", "github.com/romeufcrosa/where-to-eat"},
	})
	ctx := context.Background()
	channel := configureService(ctx)
	router := api.Router()

	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", listenAddr),
		Handler:      bugsnag.Handler(router),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go startServer(ctx, server)
	stopAtSignal(ctx, server, channel)
}

func configureService(ctx context.Context) chan os.Signal {
	googleGeo, err := domain.NewGoogleGeo()
	if err != nil {
		log.Fatal(err)
	}

	providers.Configure(
		providers.NewParams(googleGeo.Client),
		domain.IsConfigured,
	)
	providers.RegisterGatewayProviders()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	return c
}

func startServer(ctx context.Context, server *http.Server) {
	bugsnag.Notify(fmt.Errorf("Test error"))
	log.Printf("Server listening at %s\n", listenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func stopAtSignal(background context.Context, server *http.Server, stop chan os.Signal) {
	<-stop

	ctx, cancel := context.WithTimeout(background, 15*time.Second)
	defer cancel()

	log.Println("Shutting down service")
	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	close(stop)
}
