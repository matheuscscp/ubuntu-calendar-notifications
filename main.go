package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func main() {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	svc, err := calendar.NewService(context.Background(), option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		l.WithError(err).Fatal("error creating calendar client")
	}

	events, err := svc.Events.List(os.Getenv("GMAIL")).Do()
	if err != nil {
		l.WithError(err).Fatal("error listing pimenta events")
	}

	fmt.Println(events)
}
