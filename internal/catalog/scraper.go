// Package catalog, spor.istanbul kiralama sayfasından spor türü / tesis / salon
// bilgilerini çekerek yerel DB'ye kaydeder.
package catalog

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"

	"appointment-scrapper/internal/browser"
	catalogrepo "appointment-scrapper/repository/catalog"
)

const (
	baseURL    = "https://online.spor.istanbul"
	loginURL   = baseURL + "/uyegiris"
	homeURL    = baseURL + "/anasayfa"
	kiralamaJS = `
		(function(){
			var tab = document.querySelector('a#kytab');
			if(tab){ tab.click(); return 'clicked'; }
			return 'not_found';
		})()`
)

// Scraper spor.istanbul katalog verisini çeker.
type Scraper struct {
	repo   *catalogrepo.CatalogRepository
	logger *zap.Logger
}

func NewScraper(repo *catalogrepo.CatalogRepository, logger *zap.Logger) *Scraper {
	return &Scraper{repo: repo, logger: logger}
}

// Sync tarayıcıyı açar, siteye giriş yapar, tüm spor/tesis/salon verisini
// çekerek DB'ye yazar. tcNo ve password config'den gelir.
func (s *Scraper) Sync(ctx context.Context, tcNo, password string) error {
	br := browser.New(true, 90, s.logger)
	bCtx, cancel := br.NewContext(ctx)
	defer cancel()

	s.logger.Info("Siteye giriş yapılıyor...")
	if err := s.login(bCtx, tcNo, password); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	s.logger.Info("Kiralama sekmesine geçiliyor...")
	if err := chromedp.Run(bCtx,
		chromedp.Navigate(homeURL),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`a#kytab`, chromedp.ByQuery),
		chromedp.Click(`a#kytab`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	); err != nil {
		return fmt.Errorf("kiralama sekmesi: %w", err)
	}

	s.logger.Info("Spor dalları okunuyor...")
	sportOptions, err := s.getOptions(bCtx, "ddlKiralikBransFiltre")
	if err != nil {
		return fmt.Errorf("spor dalları okunamadı: %w", err)
	}
	s.logger.Info("Spor dalları bulundu", zap.Int("count", len(sportOptions)))

	var items []catalogrepo.CatalogItem

	for _, sport := range sportOptions {
		if sport.Value == "" || sport.Value == "0" {
			continue
		}
		s.logger.Info("Spor dalı işleniyor", zap.String("name", sport.Text))

		// Spor dalı seç → tesis AJAX
		if err := s.triggerSelect(bCtx, "ddlKiralikBransFiltre", sport.Value); err != nil {
			s.logger.Warn("Spor dalı seçilemedi", zap.String("name", sport.Text), zap.Error(err))
			continue
		}
		if err := s.waitForOptions(bCtx, "ddlKiralikTesisFiltre", 2, 6*time.Second); err != nil {
			s.logger.Warn("Tesis dropdown dolmadı", zap.String("sport", sport.Text))
			continue
		}

		facilityOptions, err := s.getOptions(bCtx, "ddlKiralikTesisFiltre")
		if err != nil {
			s.logger.Warn("Tesisler okunamadı", zap.Error(err))
			continue
		}
		s.logger.Info("Tesisler bulundu",
			zap.String("sport", sport.Text),
			zap.Int("count", len(facilityOptions)))

		for _, facility := range facilityOptions {
			if facility.Value == "" || facility.Value == "0" {
				continue
			}

			// Tesis seç → salon AJAX
			if err := s.triggerSelect(bCtx, "ddlKiralikTesisFiltre", facility.Value); err != nil {
				s.logger.Warn("Tesis seçilemedi", zap.String("name", facility.Text), zap.Error(err))
				// Salon olmadan da kaydet
				items = append(items, catalogrepo.CatalogItem{
					SportTypeName:      sport.Text,
					SportTypeSiteValue: sport.Value,
					FacilityName:       facility.Text,
					FacilitySiteValue:  facility.Value,
				})
				continue
			}
			if err := s.waitForOptions(bCtx, "ddlKiralikSalonFiltre", 2, 5*time.Second); err != nil {
				// Salon yoksa tesis'i yalnız kaydet
				items = append(items, catalogrepo.CatalogItem{
					SportTypeName:      sport.Text,
					SportTypeSiteValue: sport.Value,
					FacilityName:       facility.Text,
					FacilitySiteValue:  facility.Value,
				})
				continue
			}

			courtOptions, err := s.getOptions(bCtx, "ddlKiralikSalonFiltre")
			if err != nil {
				s.logger.Warn("Salonlar okunamadı", zap.Error(err))
				continue
			}

			for _, court := range courtOptions {
				if court.Value == "" || court.Value == "0" {
					continue
				}
				items = append(items, catalogrepo.CatalogItem{
					SportTypeName:      sport.Text,
					SportTypeSiteValue: sport.Value,
					FacilityName:       facility.Text,
					FacilitySiteValue:  facility.Value,
					CourtName:          court.Text,
					CourtSiteValue:     court.Value,
				})
			}
		}
	}

	s.logger.Info("DB'ye yazılıyor...", zap.Int("item_count", len(items)))
	if err := s.repo.UpsertCatalog(ctx, items); err != nil {
		return fmt.Errorf("upsert: %w", err)
	}
	s.logger.Info("Katalog sync tamamlandı", zap.Int("items", len(items)))
	return nil
}

// ─── Yardımcılar ─────────────────────────────────────────────────────────────

type selectOption struct {
	Value string
	Text  string
}

func (s *Scraper) login(ctx context.Context, tcNo, password string) error {
	return chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(`#txtTCPasaport`, chromedp.ByQuery),
		chromedp.Clear(`#txtTCPasaport`, chromedp.ByQuery),
		chromedp.SendKeys(`#txtTCPasaport`, tcNo, chromedp.ByQuery),
		chromedp.Clear(`#txtSifre`, chromedp.ByQuery),
		chromedp.SendKeys(`#txtSifre`, password, chromedp.ByQuery),
		chromedp.Click(`#btnGirisYap`, chromedp.ByQuery),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
}

// getOptions bir select'in tüm option'larını döner.
func (s *Scraper) getOptions(ctx context.Context, elemID string) ([]selectOption, error) {
	var rawValues []string
	var rawTexts []string

	script := fmt.Sprintf(`(function(){
		var sel = document.getElementById(%q);
		if(!sel) return [];
		return [...sel.options].map(o => o.value);
	})()`, elemID)

	scriptText := fmt.Sprintf(`(function(){
		var sel = document.getElementById(%q);
		if(!sel) return [];
		return [...sel.options].map(o => o.text.trim());
	})()`, elemID)

	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &rawValues)); err != nil {
		return nil, err
	}
	if err := chromedp.Run(ctx, chromedp.Evaluate(scriptText, &rawTexts)); err != nil {
		return nil, err
	}

	var opts []selectOption
	for i := range rawValues {
		if i < len(rawTexts) {
			opts = append(opts, selectOption{Value: rawValues[i], Text: rawTexts[i]})
		}
	}
	return opts, nil
}

// triggerSelect bir select element'in değerini seçer ve change event'i tetikler.
func (s *Scraper) triggerSelect(ctx context.Context, elemID, value string) error {
	script := fmt.Sprintf(`(function(){
		var sel = document.getElementById(%q);
		if(!sel) return 'NOT_FOUND';
		sel.value = %q;
		sel.dispatchEvent(new Event('change', {bubbles:true}));
		return 'OK';
	})()`, elemID, value)

	var result string
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &result)); err != nil {
		return err
	}
	if result != "OK" {
		return fmt.Errorf("select tetiklenemedi: %s → %s", elemID, result)
	}
	return nil
}

// waitForOptions bir select'in dolmasını bekler.
func (s *Scraper) waitForOptions(ctx context.Context, elemID string, minOpts int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var count int
		_ = chromedp.Run(ctx, chromedp.Evaluate(
			fmt.Sprintf(`(function(){var s=document.getElementById(%q);return s?s.options.length:0;})()`, elemID),
			&count,
		))
		if count >= minOpts {
			return nil
		}
		time.Sleep(400 * time.Millisecond)
	}
	return fmt.Errorf("timeout: %s min %d option bekleniyor", elemID, minOpts)
}
