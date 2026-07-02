package service

import (
	"context"
	"fmt"
	"time"

	"appointment-scrapper/internal/notifier"
)

// APINotifier notifier.Notifier arayüzünü uygular.
// WaitForReply çağrıldığında hem Telegram polling hem de UI'dan gelen
// SMS kodlarını dinler; hangisi önce gelirse onu döner.
type APINotifier struct {
	jobID    string
	telegram *notifier.TelegramNotifier
	registry *SMSCodeRegistry
}

func NewAPINotifier(jobID string, tg *notifier.TelegramNotifier, registry *SMSCodeRegistry) *APINotifier {
	return &APINotifier{
		jobID:    jobID,
		telegram: tg,
		registry: registry,
	}
}

func (n *APINotifier) Send(ctx context.Context, msg string) error {
	if n.telegram != nil {
		return n.telegram.Send(ctx, msg)
	}
	return nil
}

// WaitForReply SMS kodunu iki kanaldan yarışmalı olarak bekler:
// 1. Telegram long-polling (telegram yapılandırılmışsa)
// 2. POST /api/v1/jobs/:id/sms-reply endpoint'i üzerinden UI girişi
func (n *APINotifier) WaitForReply(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	// UI için kayıt aç
	apiCh := n.registry.Register(n.jobID)
	defer n.registry.Unregister(n.jobID)

	tgCh := make(chan string, 1)
	if n.telegram != nil {
		tgCtx, tgCancel := context.WithCancel(ctx)
		defer tgCancel()
		go func() {
			code, err := n.telegram.WaitForReply(tgCtx, prompt, timeout)
			if err == nil && code != "" {
				select {
				case tgCh <- code:
				default:
				}
			}
		}()
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case code := <-apiCh:
		return code, nil
	case code := <-tgCh:
		return code, nil
	case <-timer.C:
		return "", fmt.Errorf("SMS kodu bekleme zaman aşımı (%v)", timeout)
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
