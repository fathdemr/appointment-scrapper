package browser

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"

	"appointment-scrapper/config"
)

type Browser struct {
	cfg    config.ScraperConfig
	logger *zap.Logger
}

func New(cfg config.ScraperConfig, logger *zap.Logger) *Browser {
	return &Browser{cfg: cfg, logger: logger}
}

// NewContext chromedp context'i headless/headful modda döner.
func (b *Browser) NewContext(parent context.Context) (context.Context, context.CancelFunc) {
	opts := chromedp.DefaultExecAllocatorOptions[:]

	if !b.cfg.Headless {
		opts = append(opts,
			chromedp.Flag("headless", false),
			chromedp.Flag("disable-gpu", false),
		)
	}

	opts = append(opts,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent(
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "+
				"AppleWebKit/537.36 (KHTML, like Gecko) "+
				"Chrome/120.0.0.0 Safari/537.36",
		),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, opts...)
	timeout := time.Duration(b.cfg.BrowserTimeoutSeconds) * time.Second
	ctx, ctxCancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(format string, args ...interface{}) {
			b.logger.Sugar().Debugf(format, args...)
		}),
	)
	ctx, timeoutCancel := context.WithTimeout(ctx, timeout)

	cancel := func() {
		timeoutCancel()
		ctxCancel()
		allocCancel()
	}

	return ctx, cancel
}
