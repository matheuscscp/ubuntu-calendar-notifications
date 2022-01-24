package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var (
	logger    logrus.FieldLogger
	startTime time.Time
)

func init() {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	logger = l
	startTime = time.Now()
}

type client struct {
	*calendar.Service
}

func (c *client) getTodayForGmail(ctx context.Context, gmail string) ([]*calendar.Event, error) {
	t0 := time.Now()
	beg := time.Date(t0.Year(), t0.Month(), t0.Day(), 0, 0, 0, 0, t0.Location())
	end := beg.Add(24 * time.Hour)
	events, err := c.Events.List(gmail).
		TimeMin(beg.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}
	return events.Items, nil
}

func (c *client) getTodayForGmails(ctx context.Context, gmails []string) []*calendar.Event {
	var events []*calendar.Event
	for _, gmail := range gmails {
		e, err := c.getTodayForGmail(ctx, gmail)
		if err != nil {
			logger.
				WithError(err).
				WithField("calendar-id", gmail).
				Error("error listing events for calendar")
		} else {
			events = append(events, e...)
		}
	}
	return events
}

func mustParseTime(s string) time.Time {
	x, _ := time.Parse(time.RFC3339, s)
	return x
}

func notify(events []*calendar.Event) error {
	now := time.Now()
	for _, e := range events {
		if e == nil {
			logger.Warn("nil event")
			continue
		}
		if e.Start == nil || e.End == nil {
			logger.WithField("event", e).Warn("event with nil start or end")
		}
		beg := mustParseTime(e.Start.DateTime)
		end := mustParseTime(e.End.DateTime)
		if startTime.Before(beg) && beg.Before(now) && now.Before(end) {
			summary := fmt.Sprintf("%s - %s", beg.Format(time.Kitchen), end.Format(time.Kitchen))
			body := e.Summary
			if body == "" {
				body = fmt.Sprintf("Event ID: %s", e.Id)
			}
			if err := exec.Command(
				"notify-send",
				"--urgency=critical",
				summary,
				body,
			).Run(); err != nil {
				logger.
					WithError(err).
					WithField("summary", summary).
					WithField("body", body).
					Error("error sending ubuntu notification")
			}
		}
	}
	return nil
}

func runWithCancelOnSignal(sigch <-chan os.Signal, cb func(context.Context) error) (os.Signal, error) {
	var sig os.Signal

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		defer cancel()

		select {
		case <-ctx.Done():
		case sig = <-sigch:
		}
	}()
	err := cb(ctx)

	if sig == nil {
		return nil, err
	}

	if err != nil && strings.Contains(err.Error(), "context canceled") {
		return sig, nil
	}

	return sig, err
}

func main() {
	// listen interruption signals
	sigch := make(chan os.Signal, 1)
	defer close(sigch)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	logTerminate := func(sig os.Signal, where string) {
		logger.
			WithField("signal", sig.String()).
			Infof("signal received during %s, terminating", where)
	}

	// connect to calendar api
	var svc *calendar.Service
	sig, err := runWithCancelOnSignal(sigch, func(ctx context.Context) error {
		var err error
		svc, err = calendar.NewService(
			ctx,
			option.WithCredentialsFile(os.Getenv("CREDENTIALS_FILE")),
		)
		return err
	})
	if sig != nil {
		logTerminate(sig, "connect")
		return
	}
	if err != nil {
		logger.WithError(err).Fatal("error creating calendar client")
	}
	client := &client{Service: svc}

	// close api call over configs
	gmails := strings.Split(os.Getenv("GMAILS"), ",")
	list := func(ctx context.Context) []*calendar.Event {
		return client.getTodayForGmails(ctx, gmails)
	}

	// poll loop
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	for {
		var events []*calendar.Event
		sig, _ := runWithCancelOnSignal(sigch, func(ctx context.Context) error {
			events = list(ctx)
			return nil
		})
		if sig != nil {
			logTerminate(sig, "poll")
			return
		}
		notify(events)

		// sleep
		timer.Reset(time.Minute)
		select {
		case <-timer.C:
		case s := <-sigch:
			logTerminate(s, "sleep")
			return
		}
	}
}
