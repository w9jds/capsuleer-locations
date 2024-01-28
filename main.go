package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	esi "github.com/w9jds/go.esi"

	"cloud.google.com/go/logging"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"google.golang.org/api/option"
)

var (
	ctx        = context.Background()
	logClient  *logging.Client
	httpClient *http.Client
	esiClient  *esi.Client
	database   *db.Client
)

type CustomTransport struct {
	tripper http.RoundTripper
}

func (transport CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "Location Manager - Chingy Chonga/Jeremy Shore - w9jds@live.com")
	return transport.tripper.RoundTrip(req)
}

func main() {
	var err error
	opt := option.WithCredentialsFile("./config/new-eden-admin.json")
	// logClient, err = logging.NewClient(ctx, "new-eden-storage-a5c23")
	// if err != nil {
	// 	log.Fatalf("Failed to create logging client: %v", err)
	// }

	defer logClient.Close()

	config := &firebase.Config{
		ProjectID:   "new-eden-storage-a5c23",
		DatabaseURL: "https://new-eden-storage-a5c23.firebaseio.com",
	}

	httpClient = &http.Client{
		Transport: &CustomTransport{tripper: http.DefaultTransport},
	}
	esiClient = esi.CreateClient(httpClient)

	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("Error initializing firebase app: %v", err)
		return
	}

	database, err = app.Database(ctx)
	if err != nil {
		log.Fatalf("Error fetching firebase client: %v", err)
		return
	}

	start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
