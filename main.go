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
		return nil, fmt.Errorf(`error listing "%s" events: %w`, gmail, err)
	}
	return events.Items, nil
}

func (c *client) getTodayForGmails(ctx context.Context, gmails []string) ([]*calendar.Event, error) {
	var events []*calendar.Event
	for _, gmail := range gmails {
		e, err := c.getTodayForGmail(ctx, gmail)
		if err != nil {
			return nil, err
		}
		events = append(events, e...)
	}
	return events, nil
}

func mustParseTime(s string) time.Time {
	x, _ := time.Parse(time.RFC3339, s)
	return x
}

func notify(events []*calendar.Event) error {
	now := time.Now()
	for _, e := range events {
		beg := mustParseTime(e.Start.DateTime)
		end := mustParseTime(e.End.DateTime)
		if startTime.Before(beg) && beg.Before(now) && now.Before(end) {
			name := e.Id
			if e.Summary != "" {
				name = e.Summary
			}
			notification := fmt.Sprintf("%s (%s - %s)",
				name,
				beg.Format(time.Kitchen),
				end.Format(time.Kitchen),
			)
			if err := exec.Command("notify-send", notification).Run(); err != nil {
				logger.
					WithError(err).
					WithField("notification", notification).
					Error("error sending notification")
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
		select {
		case <-ctx.Done():
		case s := <-sigch:
			sig = s
		}
		cancel()
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

	// close api call over configs
	client := &client{Service: svc}
	gmails := strings.Split(os.Getenv("GMAILS"), ",")
	list := func(ctx context.Context) ([]*calendar.Event, error) {
		events, err := client.getTodayForGmails(ctx, gmails)
		if err != nil {
			return nil, fmt.Errorf("error listing today events: %w", err)
		}
		return events, nil
	}

	// poll loop
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	for {
		var events []*calendar.Event
		sig, err := runWithCancelOnSignal(sigch, func(ctx context.Context) error {
			var err error
			events, err = list(ctx)
			return err
		})
		if sig != nil {
			logTerminate(sig, "poll")
			return
		}
		if err != nil {
			logger.WithError(err).Error("error listing events")
		} else {
			notify(events)
		}

		timer.Reset(time.Minute)
		select {
		case <-timer.C:
		case s := <-sigch:
			logTerminate(s, "sleep")
			return
		}
	}
}
