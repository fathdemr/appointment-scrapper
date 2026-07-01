package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"

	"appointment-scrapper/config"
	"appointment-scrapper/internal/browser"
	"appointment-scrapper/internal/notifier"
)

const (
	baseURL  = "https://online.spor.istanbul"
	loginURL = baseURL + "/uyegiris"
	homeURL  = baseURL + "/anasayfa"
	cartURL  = baseURL + "/sepet"
)

// Gerçek site DOM'undan keşfedilen selector'lar
const (
	selTCInput  = `#txtTCPasaport`
	selPwdInput = `#txtSifre`
	selLoginBtn = `#btnGirisYap`

	selKiralamaTab = `a#kytab`

	// Kiralama kademeli dropdown'lar (AJAX: spor → tesis → salon)
	selSportDrop = `#ddlKiralikBransFiltre`
	selSearchBtn = `#pageContent_ucUrunArama_lbtnKiralikAra`
)

type Scraper struct {
	cfg      config.ScraperConfig
	creds    config.CredentialsConfig
	browser  *browser.Browser
	notifier notifier.Notifier
	logger   *zap.Logger
}

func New(
	cfg config.ScraperConfig,
	creds config.CredentialsConfig,
	br *browser.Browser,
	n notifier.Notifier,
	logger *zap.Logger,
) *Scraper {
	return &Scraper{
		cfg:      cfg,
		creds:    creds,
		browser:  br,
		notifier: n,
		logger:   logger,
	}
}

// Run bir giriş + arama + sepete ekleme denemesi yapar.
func (s *Scraper) Run(ctx context.Context) (bool, error) {
	bCtx, cancel := s.browser.NewContext(ctx)
	defer cancel()

	s.logger.Info("Giriş yapılıyor...")
	if err := s.login(bCtx); err != nil {
		return false, fmt.Errorf("login: %w", err)
	}
	s.logger.Info("Giriş başarılı")

	facilities := s.cfg.Facilities
	if len(facilities) == 0 {
		facilities = []string{""}
	}
	courts := s.cfg.Courts
	if len(courts) == 0 {
		courts = []string{""}
	}

	for _, facility := range facilities {
		for _, court := range courts {
			s.logger.Info("Aranıyor", zap.String("facility", facility), zap.String("court", court))
			found, result, err := s.findAndBook(bCtx, facility, court)
			if err != nil {
				s.logger.Warn("Arama hatası",
					zap.String("facility", facility),
					zap.String("court", court),
					zap.Error(err),
				)
				continue
			}
			if found {
				msg := notifier.FormatBookingMessage(*result, "Randevu sepette 30 dakika içinde ödeme yapılmazsa düşer!")
				if err := s.notifier.Send(ctx, msg); err != nil {
					s.logger.Warn("Bildirim gönderilemedi", zap.Error(err))
				}
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *Scraper) login(ctx context.Context) error {
	return chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(selTCInput, chromedp.ByQuery),
		chromedp.Clear(selTCInput, chromedp.ByQuery),
		chromedp.SendKeys(selTCInput, s.creds.TCNo, chromedp.ByQuery),
		chromedp.Clear(selPwdInput, chromedp.ByQuery),
		chromedp.SendKeys(selPwdInput, s.creds.Password, chromedp.ByQuery),
		chromedp.Click(selLoginBtn, chromedp.ByQuery),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
}

func (s *Scraper) findAndBook(ctx context.Context, facility, court string) (bool, *notifier.BookingResult, error) {
	// Ana sayfaya git ve kiralama tab'ını aç
	if err := chromedp.Run(ctx,
		chromedp.Navigate(homeURL),
		chromedp.WaitVisible(selKiralamaTab, chromedp.ByQuery),
		chromedp.Click(selKiralamaTab, chromedp.ByQuery),
		chromedp.Sleep(time.Second),
	); err != nil {
		return false, nil, fmt.Errorf("kiralama tab: %w", err)
	}

	// 1. Spor dalı seç (AJAX tetikler → tesis dropdown dolar)
	if s.cfg.SportType != "" {
		var result string
		if err := chromedp.Run(ctx,
			chromedp.WaitVisible(selSportDrop, chromedp.ByQuery),
			chromedp.EvaluateAsDevTools(
				fmt.Sprintf(`(function(){
					var search=%q.toUpperCase();
					var sel=document.getElementById('ddlKiralikBransFiltre');
					if(!sel) return 'NOT_FOUND';
					var opt=[...sel.options].find(o=>o.text.trim().toUpperCase().includes(search));
					if(!opt) return 'NO_MATCH';
					sel.value=opt.value;
					sel.dispatchEvent(new Event('change',{bubbles:true}));
					return 'OK:'+opt.value;
				})()`, s.cfg.SportType),
				&result,
			),
		); err != nil {
			s.logger.Warn("Spor dalı seçilemedi", zap.Error(err))
		} else {
			s.logger.Info("Spor dalı seçildi", zap.String("result", result))
		}
		// AJAX'ın tesis dropdown'ını doldurmasını bekle
		if err := s.waitForOptions(ctx, "ddlKiralikTesisFiltre", 2, 5*time.Second); err != nil {
			s.logger.Warn("Tesis dropdown dolmadı", zap.Error(err))
		}
	}

	// 2. Tesis seç (AJAX tetikler → salon/kort dropdown dolar)
	if facility != "" {
		var result string
		if err := chromedp.Run(ctx,
			chromedp.EvaluateAsDevTools(
				fmt.Sprintf(`(function(){
					var search=%q.toUpperCase();
					var sel=document.getElementById('ddlKiralikTesisFiltre');
					if(!sel) return 'NOT_FOUND';
					var opt=[...sel.options].find(o=>o.text.trim().toUpperCase().includes(search));
					if(!opt) return 'NO_MATCH:'+[...sel.options].map(o=>o.text).join('|');
					sel.value=opt.value;
					sel.dispatchEvent(new Event('change',{bubbles:true}));
					return 'OK:'+opt.text;
				})()`, facility),
				&result,
			),
		); err != nil {
			s.logger.Warn("Tesis seçilemedi", zap.Error(err))
		} else {
			s.logger.Info("Tesis seçildi", zap.String("result", result))
		}
		// Salon dropdown'ının dolmasını bekle
		if err := s.waitForOptions(ctx, "ddlKiralikSalonFiltre", 2, 5*time.Second); err != nil {
			s.logger.Warn("Salon dropdown dolmadı", zap.Error(err))
		}
	}

	// 3. Salon/Kort seç
	if court != "" {
		var result string
		_ = chromedp.Run(ctx,
			chromedp.EvaluateAsDevTools(
				fmt.Sprintf(`(function(){
					var search=%q.toUpperCase();
					var sel=document.getElementById('ddlKiralikSalonFiltre');
					if(!sel) return 'NOT_FOUND';
					var opt=[...sel.options].find(o=>o.text.trim().toUpperCase().includes(search));
					if(!opt) return 'NO_MATCH:'+[...sel.options].map(o=>o.text).join('|');
					sel.value=opt.value;
					sel.dispatchEvent(new Event('change',{bubbles:true}));
					return 'OK:'+opt.text;
				})()`, court),
				&result,
			),
		)
		s.logger.Info("Salon seçildi", zap.String("result", result))
	}

	// 4. Ara butonuna tıkla
	if err := chromedp.Run(ctx,
		chromedp.Click(selSearchBtn, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	); err != nil {
		return false, nil, fmt.Errorf("ara butonu: %w", err)
	}

	return s.parseAndSelectSlot(ctx, facility, court)
}

// waitForOptions bir select'in en az minOptions option içermesini timeout süresi boyunca bekler.
func (s *Scraper) waitForOptions(ctx context.Context, elemID string, minOptions int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var hasOpts bool
		_ = chromedp.Run(ctx, chromedp.Evaluate(
			fmt.Sprintf(`(function(){var s=document.getElementById(%q);return s&&s.options.length>=%d;})()`, elemID, minOptions),
			&hasOpts,
		))
		if hasOpts {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("timeout: %s min %d option bekleniyor", elemID, minOptions)
}

// parseAndSelectSlot hedef tarih+saat kombinasyonunu bulur ve
// tam akışı yürütür: Rezervasyon → Sepete Ekle → SMS doğrulama.
// Yapı: rptList outer=günler, h3.panel-title içinde tarih, inner=saat slotları.
// lblSeans_N → lbRezervasyon_N (aynı index).
func (s *Scraper) parseAndSelectSlot(ctx context.Context, facility, court string) (bool, *notifier.BookingResult, error) {
	for _, targetDate := range s.cfg.TargetDates {
		for _, desiredTime := range s.cfg.DesiredTimes {
			var btnID string
			_ = chromedp.Run(ctx, chromedp.Evaluate(
				fmt.Sprintf(`(function(){
				var targetDate = %q;  // "04.07.2026"
				var targetTime = %q;  // "21:00"

				// O tarihin gün panelini bul (h3.panel-title içinde tarih geçiyor)
				var h3s = document.querySelectorAll('h3.panel-title');
				var dayPanel = null;
				for(var h3 of h3s){
					if(h3.textContent.includes(targetDate)){
						dayPanel = h3.closest('.panel') || h3.parentElement;
						while(dayPanel && !dayPanel.querySelector('span.lblStyle')){
							dayPanel = dayPanel.parentElement;
						}
						break;
					}
				}
				if(!dayPanel) return 'DATE_NOT_FOUND:'+targetDate;

				// Panel içindeki saat span'larında hedef saati ara
				var spans = dayPanel.querySelectorAll('span.lblStyle');
				for(var sp of spans){
					if(!sp.textContent.trim().startsWith(targetTime)) continue;
					// ID: pageContent_rptList_rpChild_5_lblSeans_9
					// → pageContent_rptList_rpChild_5_lbRezervasyon_9
					var btnId = sp.id.replace('lblSeans_', 'lbRezervasyon_');
					var btn = document.getElementById(btnId);
					if(btn) return btn.id;
					return 'BTN_NOT_FOUND:'+btnId;
				}
				return 'TIME_NOT_FOUND:'+targetTime;
			})()`, targetDate, desiredTime),
				&btnID,
			))

			if btnID == "" || btnID == "DATE_NOT_FOUND:"+targetDate || btnID == "TIME_NOT_FOUND:"+desiredTime {
				s.logger.Debug("Slot bulunamadı",
					zap.String("date", targetDate),
					zap.String("time", desiredTime),
					zap.String("reason", btnID),
				)
				continue
			}

			s.logger.Info("Slot bulundu",
				zap.String("date", targetDate),
				zap.String("time", desiredTime),
				zap.String("btn", btnID),
			)

			// Adım 1: "Rezervasyon" → alert bypass
			if err := chromedp.Run(ctx,
				chromedp.Evaluate(`window.alert = function(){ return true; };`, nil),
				chromedp.Click(`a#`+btnID, chromedp.ByQuery),
				chromedp.Sleep(2*time.Second),
			); err != nil {
				s.logger.Warn("Rezervasyon tıklanamadı", zap.Error(err))
				continue
			}

			// Adım 2: "Sepete Ekle" görünür mü?
			var sepeteVisible bool
			_ = chromedp.Run(ctx, chromedp.Evaluate(
				`(function(){var b=document.getElementById('pageContent_lbtnSepeteEkle');return b&&b.offsetParent!==null;})()`,
				&sepeteVisible,
			))
			if !sepeteVisible {
				s.logger.Warn("Sepete Ekle butonu görünmedi")
				continue
			}

			s.logger.Info("Sepete Ekle tıklanıyor...")
			if err := chromedp.Run(ctx,
				chromedp.Click(`#pageContent_lbtnSepeteEkle`, chromedp.ByQuery),
				chromedp.Sleep(2*time.Second),
			); err != nil {
				s.logger.Warn("Sepete Ekle tıklanamadı", zap.Error(err))
				continue
			}

			// Adım 3: SMS ekranı var mı?
			var smsVisible bool
			_ = chromedp.Run(ctx, chromedp.Evaluate(
				`(function(){var inp=document.getElementById('pageContent_txtDogrulamaKodu');return inp&&inp.offsetParent!==null;})()`,
				&smsVisible,
			))

			if smsVisible {
				s.logger.Info("SMS doğrulama gerekli, Telegram'dan kod bekleniyor...")
				smsCtx, smsCancel := context.WithTimeout(ctx, 3*time.Minute)
				code, err := s.notifier.WaitForReply(
					smsCtx,
					"📱 <b>SMS doğrulama kodu gönderildi!</b>\n\nTelefonuna gelen <b>6 haneli kodu</b> buraya yaz:",
					3*time.Minute,
				)
				smsCancel()

				if err != nil || code == "" {
					return false, nil, fmt.Errorf("SMS kodu alınamadı: %w", err)
				}

				s.logger.Info("SMS kodu giriliyor...", zap.String("code", code))
				if err := chromedp.Run(ctx,
					chromedp.Clear(`#pageContent_txtDogrulamaKodu`, chromedp.ByQuery),
					chromedp.SendKeys(`#pageContent_txtDogrulamaKodu`, code, chromedp.ByQuery),
					chromedp.Click(`#btnCepTelDogrulamaGonder`, chromedp.ByQuery),
					chromedp.Sleep(3*time.Second),
				); err != nil {
					return false, nil, fmt.Errorf("SMS onaylama hatası: %w", err)
				}
				s.logger.Info("SMS doğrulama tamamlandı")
			}

			return true, &notifier.BookingResult{
				Facility:  facility,
				SportType: s.cfg.SportType,
				Court:     court,
				Date:      targetDate,
				Time:      desiredTime,
				CartURL:   cartURL,
			}, nil
		} // end desiredTime loop
	} // end targetDate loop

	return false, nil, nil
}
