package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	esi "github.com/w9jds/go.esi"
	"google.golang.org/api/option"
)

var (
	ctx        = context.Background()
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
	opt := option.WithCredentialsFile("./config/new-eden-admin.json")
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
		log.Println("Error initializing firebase app: ", err)
		return
	}

	database, err = app.Database(ctx)
	if err != nil {
		log.Println("Error fetching firebase client: ", err)
		return
	}

	start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
