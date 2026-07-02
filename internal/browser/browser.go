package browser

import (
	"context"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

type Browser struct {
	headless bool
	timeout  time.Duration
	logger   *zap.Logger
}

func New(headless bool, timeoutSeconds int, logger *zap.Logger) *Browser {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	return &Browser{
		headless: headless,
		timeout:  time.Duration(timeoutSeconds) * time.Second,
		logger:   logger,
	}
}

func (b *Browser) NewContext(parent context.Context) (context.Context, context.CancelFunc) {
	// DefaultExecAllocatorOptions yerine sıfırdan başla — container'da çakışma yaratabilir.
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	}

	// Container ortamında (Alpine/K8s) CHROME_PATH env var ile binary yolu belirtilir.
	if p := os.Getenv("CHROME_PATH"); p != "" {
		opts = append(opts, chromedp.ExecPath(p))
	}

	if b.headless {
		opts = append(opts, chromedp.Headless)
	}

	opts = append(opts,
		// Sandbox
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),

		// /dev/shm — K8s varsayılan 64MB, yetmez
		chromedp.Flag("disable-dev-shm-usage", true),

		// GPU
		chromedp.DisableGPU,
		chromedp.Flag("disable-software-rasterizer", true),

		// Crashpad subprocess'i tamamen kapat.
		// "chrome_crashpad_handler: --database is required" bu flagsiz gelir.
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("single-process", true), // subprocess açamayan K8s ortamları için kesin çözüm
		chromedp.Flag("disable-crash-reporter", true),

		// Chrome'a yazılabilir bir kullanıcı dizini ver
		chromedp.Flag("user-data-dir", "/tmp/chrome-user-data"),

		// Gereksiz arka plan servisleri
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("metrics-recording-only", true),

		chromedp.UserAgent(
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "+
				"AppleWebKit/537.36 (KHTML, like Gecko) "+
				"Chrome/120.0.0.0 Safari/537.36",
		),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, opts...)
	ctx, ctxCancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(format string, args ...interface{}) {
			b.logger.Sugar().Debugf(format, args...)
		}),
	)
	ctx, timeoutCancel := context.WithTimeout(ctx, b.timeout)

	cancel := func() {
		timeoutCancel()
		ctxCancel()
		allocCancel()
	}

	return ctx, cancel
}
