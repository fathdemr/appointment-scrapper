package notifier

import (
	"context"
	"time"
)

type Notifier interface {
	Send(ctx context.Context, msg string) error
	WaitForReply(ctx context.Context, prompt string, timeout time.Duration) (string, error)
}

type BookingResult struct {
	Facility  string
	SportType string
	Court     string
	Date      string
	Time      string
	CartURL   string
}
