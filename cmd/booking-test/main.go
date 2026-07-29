// booking-test: Local test aracı — Chrome'u AÇIK tutar, TEK SEFER login olur,
// sonra sadece arama+rezervasyon adımını hızlı döngüde spamler.
//
// Kullanım:
//
//	CREDENTIALS_TC_NO=... CREDENTIALS_PASSWORD=... go run ./cmd/booking-test
//
// veya proje kökünde .env dosyası varsa (CREDENTIALS_TC_NO / CREDENTIALS_PASSWORD)
// oradan okur. Varsayılan hedef: Haliç Spor Parkı, Futbol (halı saha),
// bir sonraki Cumartesi 21:45.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"

	"appointment-scrapper/config"
)

const (
	baseURL  = "https://online.spor.istanbul"
	loginURL = baseURL + "/uyegiris"
	homeURL  = baseURL + "/anasayfa"
	cartURL  = baseURL + "/sepet"

	selTCInput  = `#txtTCPasaport`
	selPwdInput = `#txtSifre`
	selLoginBtn = `#btnGirisYap`

	selKiralamaTab = `a#kytab`
	selSportDrop   = `#ddlKiralikBransFiltre`
	selSearchBtn   = `#pageContent_ucUrunArama_lbtnKiralikAra`
)

type target struct {
	facility string
	sport    string
	court    string
	date     string // 02.01.2006
	timeStr  string // 15:04
}

func main() {
	facility := flag.String("facility", "Haliç", "Tesis adı (kısmi eşleşme)")
	sport := flag.String("sport", "Futbol", "Spor dalı (kısmi eşleşme)")
	court := flag.String("court", "", "Salon/kort adı (boş = seçme)")
	date := flag.String("date", "", "Hedef tarih 02.01.2006 (boş = bir sonraki Cumartesi)")
	timeStr := flag.String("time", "21:45", "Hedef saat HH:MM")
	interval := flag.Duration("interval", 1500*time.Millisecond, "Denemeler arası bekleme")
	flag.Parse()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	loadDotEnv(".env")
	tcNo := os.Getenv("CREDENTIALS_TC_NO")
	password := os.Getenv("CREDENTIALS_PASSWORD")
	if tcNo == "" || password == "" {
		// Env'de yoksa config.yaml'daki credentials bölümünden al
		if cfg, err := config.Load(); err == nil {
			tcNo = cfg.Credentials.TCNo
			password = cfg.Credentials.Password
		}
	}
	if tcNo == "" || password == "" {
		logger.Fatal("Giriş bilgisi bulunamadı: config.yaml credentials bölümü veya CREDENTIALS_TC_NO/CREDENTIALS_PASSWORD env gerekli")
	}

	if *date == "" {
		*date = nextSaturday(time.Now()).Format("02.01.2006")
	}
	tgt := target{
		facility: *facility,
		sport:    *sport,
		court:    *court,
		date:     *date,
		timeStr:  *timeStr,
	}

	logger.Info("Hedef",
		zap.String("tesis", tgt.facility),
		zap.String("brans", tgt.sport),
		zap.String("tarih", tgt.date),
		zap.String("saat", tgt.timeStr),
		zap.Duration("interval", *interval),
	)

	// Ctrl+C ile temiz kapanış
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ── Chrome: HEADFUL, tüm oturum boyunca tek context, global timeout YOK ──
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("user-data-dir", "/tmp/chrome-booking-test"),
		// NOT: UserAgent override YOK — gerçek (güncel) Chrome UA'sı Cloudflare
		// için daha güvenli; eski sürüm spoof'u challenge tetikleyebiliyor.
	}
	if proxy := os.Getenv("CHROME_PROXY"); proxy != "" {
		opts = append(opts, chromedp.ProxyServer(proxy))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(rootCtx, opts...)
	defer allocCancel()
	ctx, ctxCancel := chromedp.NewContext(allocCtx)
	defer ctxCancel()

	b := &bot{ctx: ctx, logger: logger, tcNo: tcNo, password: password, tgt: tgt}

	// Browser'ı uzun ömürlü ctx ile başlat — ilk Run'a timeout'lu bir context
	// verilirse chromedp browser'ın ömrünü ona bağlıyor ve timeout cancel
	// olunca Chrome kapanıyor.
	// logger.Fatal kullanma: os.Exit defer'ları atlar ve Chrome yetim kalıp
	// user-data-dir kilidini tutmaya devam eder. Error + return ile çık.
	logger.Info("Chrome başlatılıyor...")
	if err := chromedp.Run(ctx); err != nil {
		logger.Error("Chrome başlatılamadı", zap.Error(err))
		return
	}

	if err := b.login(); err != nil {
		logger.Error("Login başarısız", zap.Error(err))
		return
	}
	if err := b.setupSearchPage(); err != nil {
		logger.Error("Arama sayfası hazırlanamadı", zap.Error(err))
		return
	}

	// ── Spam döngüsü: sadece Ara + slot kontrolü ──
	attempt := 0
	for {
		select {
		case <-rootCtx.Done():
			logger.Info("Durduruldu")
			return
		default:
		}

		attempt++
		booked, err := b.attempt(attempt)
		if err != nil {
			logger.Warn("Deneme hatası, sayfa yeniden hazırlanıyor", zap.Int("attempt", attempt), zap.Error(err))
			if rerr := b.recover(); rerr != nil {
				logger.Error("Kurtarma başarısız, 5sn sonra tekrar", zap.Error(rerr))
				sleepCtx(rootCtx, 5*time.Second)
			}
			continue
		}
		if booked {
			logger.Info("✅ RANDEVU SEPETE EKLENDİ! Ödeme için 30 dakikan var.",
				zap.String("sepet", cartURL))
			fmt.Println("\nChrome açık bırakıldı — ödemeyi açık pencereden tamamlayabilirsin.")
			fmt.Println("Kapatmak için Enter'a bas...")
			waitEnter(rootCtx)
			return
		}
		sleepCtx(rootCtx, *interval)
	}
}

type bot struct {
	ctx      context.Context
	logger   *zap.Logger
	tcNo     string
	password string
	tgt      target
}

// run chromedp aksiyonlarını attempt-bazlı timeout ile çalıştırır;
// browser context'i asla öldürmez.
func (b *bot) run(timeout time.Duration, actions ...chromedp.Action) error {
	ctx, cancel := context.WithTimeout(b.ctx, timeout)
	defer cancel()
	return chromedp.Run(ctx, actions...)
}

func (b *bot) login() error {
	b.logger.Info("Giriş yapılıyor...")
	if err := b.run(30*time.Second, chromedp.Navigate(loginURL)); err != nil {
		return fmt.Errorf("login sayfası açılamadı: %w", err)
	}

	// Zaten login'liysek site anasayfaya yönlendirir
	var url string
	_ = b.run(5*time.Second, chromedp.Location(&url))
	if !strings.Contains(url, "uyegiris") {
		b.logger.Info("Zaten oturum açık (user-data-dir cookie'leri)", zap.String("url", url))
		return nil
	}

	// Alanları JS ile set et: SendKeys, autofill/maskeleme ile yarışıp değeri
	// çiftleyebiliyor (ekran görüntüsünde TC iki kez yazılmıştı).
	var fieldsOK bool
	if err := b.run(20*time.Second,
		chromedp.WaitVisible(selTCInput, chromedp.ByQuery),
		chromedp.Evaluate(fmt.Sprintf(`(function(){
			function set(id,val){
				var el=document.getElementById(id);
				if(!el) return false;
				el.focus(); el.value='';
				el.value=val;
				el.dispatchEvent(new Event('input',{bubbles:true}));
				el.dispatchEvent(new Event('change',{bubbles:true}));
				el.blur();
				return el.value===val;
			}
			return set('txtTCPasaport',%q) && set('txtSifre',%q);
		})()`, b.tcNo, b.password), &fieldsOK),
	); err != nil {
		return fmt.Errorf("login formu doldurulamadı: %w", err)
	}
	if !fieldsOK {
		return fmt.Errorf("login alanları beklenen değeri almadı")
	}
	if err := b.run(10*time.Second, chromedp.Click(selLoginBtn, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("giriş butonu: %w", err)
	}

	// Postback süresi değişken — URL login sayfasından ayrılana kadar bekle.
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		_ = b.run(3*time.Second, chromedp.Location(&url))
		if url != "" && !strings.Contains(url, "uyegiris") {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if strings.Contains(url, "uyegiris") {
		var bodyText string
		var shot []byte
		_ = b.run(8*time.Second,
			chromedp.Evaluate(`document.body ? document.body.innerText.slice(0, 1200) : ''`, &bodyText),
			chromedp.CaptureScreenshot(&shot),
		)
		if len(shot) > 0 {
			_ = os.WriteFile("/tmp/booking-test-login-fail.png", shot, 0o644)
		}
		b.logger.Error("Login tanılaması",
			zap.String("body", bodyText),
			zap.String("screenshot", "/tmp/booking-test-login-fail.png"),
		)
		return fmt.Errorf("login sonrası hâlâ giriş sayfasındayız (url=%s)", url)
	}
	b.logger.Info("Giriş başarılı", zap.String("url", url))
	return nil
}

// setupSearchPage anasayfa → kiralama tabı → branş/tesis/salon dropdown seçimi.
// Bir kez çalışır; ASP.NET postback'i seçimleri koruduğu için spam döngüsünde
// tekrar gerekmiyor.
func (b *bot) setupSearchPage() error {
	b.logger.Info("Arama sayfası hazırlanıyor...")
	if err := b.run(30*time.Second,
		chromedp.Navigate(homeURL),
		chromedp.WaitVisible(selKiralamaTab, chromedp.ByQuery),
		chromedp.Click(selKiralamaTab, chromedp.ByQuery),
		chromedp.Sleep(time.Second),
	); err != nil {
		return fmt.Errorf("kiralama tabı: %w", err)
	}

	// Branş seç → AJAX tesis dropdown'ını doldurur
	if err := b.selectDropdown("ddlKiralikBransFiltre", b.tgt.sport); err != nil {
		return fmt.Errorf("branş: %w", err)
	}
	if err := b.waitForOptions("ddlKiralikTesisFiltre", 2, 8*time.Second); err != nil {
		return fmt.Errorf("tesis dropdown dolmadı: %w", err)
	}

	// Tesis seç → AJAX salon dropdown'ını doldurur
	if err := b.selectDropdown("ddlKiralikTesisFiltre", b.tgt.facility); err != nil {
		return fmt.Errorf("tesis: %w", err)
	}

	if b.tgt.court != "" {
		if err := b.waitForOptions("ddlKiralikSalonFiltre", 2, 8*time.Second); err != nil {
			b.logger.Warn("Salon dropdown dolmadı", zap.Error(err))
		} else if err := b.selectDropdown("ddlKiralikSalonFiltre", b.tgt.court); err != nil {
			b.logger.Warn("Salon seçilemedi", zap.Error(err))
		}
	}

	b.logger.Info("Arama sayfası hazır")
	return nil
}

func (b *bot) selectDropdown(elemID, search string) error {
	var result string
	err := b.run(10*time.Second, chromedp.EvaluateAsDevTools(
		fmt.Sprintf(`(function(){
			var search=%q.toLocaleUpperCase('tr');
			var sel=document.getElementById(%q);
			if(!sel) return 'NOT_FOUND';
			var opt=[...sel.options].find(o=>o.text.trim().toLocaleUpperCase('tr').includes(search));
			if(!opt) return 'NO_MATCH:'+[...sel.options].map(o=>o.text).join('|');
			sel.value=opt.value;
			sel.dispatchEvent(new Event('change',{bubbles:true}));
			return 'OK:'+opt.text.trim();
		})()`, search, elemID),
		&result,
	))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(result, "OK:") {
		return fmt.Errorf("%s için %q seçilemedi: %s", elemID, search, result)
	}
	b.logger.Info("Seçildi", zap.String("dropdown", elemID), zap.String("secim", result))
	return nil
}

func (b *bot) waitForOptions(elemID string, minOptions int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var hasOpts bool
		_ = b.run(3*time.Second, chromedp.Evaluate(
			fmt.Sprintf(`(function(){var s=document.getElementById(%q);return !!s&&s.options.length>=%d;})()`, elemID, minOptions),
			&hasOpts,
		))
		if hasOpts {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timeout: %s dropdown'ında min %d option yok", elemID, minOptions)
}

// attempt tek bir spam denemesi: Ara'ya bas → sonuçta hedef slotu ara →
// varsa Rezervasyon + Sepete Ekle akışını yürüt.
func (b *bot) attempt(n int) (bool, error) {
	// Oturum düşmüş mü?
	var url string
	_ = b.run(5*time.Second, chromedp.Location(&url))
	if strings.Contains(url, "uyegiris") {
		b.logger.Warn("Oturum düşmüş, yeniden giriş yapılıyor")
		if err := b.login(); err != nil {
			return false, err
		}
		if err := b.setupSearchPage(); err != nil {
			return false, err
		}
	}

	// Ara butonuna bas (postback; dropdown seçimleri korunur)
	if err := b.run(20*time.Second,
		chromedp.WaitVisible(selSearchBtn, chromedp.ByQuery),
		chromedp.Click(selSearchBtn, chromedp.ByQuery),
	); err != nil {
		return false, fmt.Errorf("ara butonu: %w", err)
	}

	// Sonuç panellerinin (gün başlıkları) render olmasını bekle — postback
	// süresi değişken, sabit sleep ilk denemelerde erken kalıyordu.
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		var hasResults bool
		_ = b.run(3*time.Second, chromedp.Evaluate(
			`document.querySelectorAll('h3.panel-title').length > 0`,
			&hasResults,
		))
		if hasResults {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}

	// Hedef gün+saat için Rezervasyon butonunu ara
	var btnID string
	if err := b.run(10*time.Second, chromedp.Evaluate(
		fmt.Sprintf(`(function(){
			var targetDate = %q, targetTime = %q;
			// Gün panelleri sıralı: h3 listesindeki index = rpChild_N index'i.
			// (27.07 Pzt = rpChild_0 ... 02.08 Paz = rpChild_6 — HTML dökümüyle doğrulandı)
			var h3s = [...document.querySelectorAll('h3.panel-title')];
			var idx = h3s.findIndex(h=>h.textContent.includes(targetDate));
			if(idx < 0){
				var seen = h3s.map(h=>h.textContent.trim().replace(/\s+/g,' ')).join(' | ');
				return 'DATE_NOT_FOUND avail=['+seen+']';
			}
			var spans = document.querySelectorAll(
				'[id^="pageContent_rptList_rpChild_'+idx+'_lblSeans_"]');
			if(spans.length === 0) return 'NO_SEANCES_FOR_DAY idx='+idx;
			for(var sp of spans){
				if(!sp.textContent.trim().startsWith(targetTime)) continue;
				var btnId = sp.id.replace('lblSeans_', 'lbRezervasyon_');
				var btn = document.getElementById(btnId);
				if(btn) return btn.id;
				return 'BTN_NOT_FOUND id='+sp.id;
			}
			var times=[...spans].map(s=>s.textContent.trim().replace(/\s+/g,' ')).join(' | ');
			return 'TIME_NOT_FOUND avail=['+times+']';
		})()`, b.tgt.date, b.tgt.timeStr),
		&btnID,
	)); err != nil {
		return false, fmt.Errorf("slot tarama: %w", err)
	}

	if btnID == "" || strings.HasPrefix(btnID, "DATE_NOT_FOUND") ||
		strings.HasPrefix(btnID, "TIME_NOT_FOUND") || strings.HasPrefix(btnID, "BTN_NOT_FOUND") ||
		strings.HasPrefix(btnID, "NO_SEANCES_FOR_DAY") {
		// BTN_NOT_FOUND = slot listede ama rezervasyon linki henüz aktif değil
		// (açılış saati gelmedi) — spamlemeye devam.
		b.logger.Info("Slot henüz yok", zap.Int("attempt", n), zap.String("durum", btnID))
		if n <= 2 && strings.HasPrefix(btnID, "DATE_NOT_FOUND") {
			b.dumpDiagnostics(n)
		}
		if strings.HasPrefix(btnID, "DAY_SPANS_AMBIGUOUS") && n <= 3 {
			var html string
			_ = b.run(10*time.Second, chromedp.Evaluate(
				`(function(){var p=document.getElementById('pageContent');return p?p.outerHTML:document.body.outerHTML;})()`,
				&html,
			))
			path := fmt.Sprintf("/tmp/booking-test-results-%d.html", n)
			_ = os.WriteFile(path, []byte(html), 0o644)
			b.logger.Warn("Sonuç HTML dökümü alındı", zap.String("path", path), zap.Int("bytes", len(html)))
		}
		return false, nil
	}

	b.logger.Info("🎯 SLOT AKTİF! Rezervasyon deneniyor...", zap.String("btn", btnID))

	// Rezervasyon → confirm/alert bypass → Sepete Ekle
	if err := b.run(20*time.Second,
		chromedp.Evaluate(`window.alert=function(){return true;};window.confirm=function(){return true;};`, nil),
		chromedp.Click(`a#`+btnID, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	); err != nil {
		return false, fmt.Errorf("rezervasyon tıklama: %w", err)
	}

	var sepeteVisible bool
	_ = b.run(5*time.Second, chromedp.Evaluate(
		`(function(){var b=document.getElementById('pageContent_lbtnSepeteEkle');return !!b&&b.offsetParent!==null;})()`,
		&sepeteVisible,
	))
	if !sepeteVisible {
		return false, fmt.Errorf("sepete ekle butonu görünmedi (slot kapılmış olabilir)")
	}

	b.logger.Info("Sepete Ekle tıklanıyor...")
	if err := b.run(15*time.Second,
		chromedp.Click(`#pageContent_lbtnSepeteEkle`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	); err != nil {
		return false, fmt.Errorf("sepete ekle: %w", err)
	}

	// SMS doğrulaması çıktıysa kodu terminalden al
	var smsVisible bool
	_ = b.run(5*time.Second, chromedp.Evaluate(
		`(function(){var i=document.getElementById('pageContent_txtDogrulamaKodu');return !!i&&i.offsetParent!==null;})()`,
		&smsVisible,
	))
	if smsVisible {
		fmt.Println("\n📱 SMS doğrulama gerekli! Kodu şu şekilde ver:")
		fmt.Println("   echo KOD > sms-code.txt   (proje kökünde)")
		fmt.Println("   veya terminale yazıp Enter'a bas")
		code := waitForSMSCode(b.ctx, 3*time.Minute)
		if code == "" {
			return false, fmt.Errorf("SMS kodu alınamadı (3dk timeout)")
		}
		b.logger.Info("SMS kodu alındı, giriliyor...")
		if err := b.run(20*time.Second,
			chromedp.Clear(`#pageContent_txtDogrulamaKodu`, chromedp.ByQuery),
			chromedp.SendKeys(`#pageContent_txtDogrulamaKodu`, code, chromedp.ByQuery),
			chromedp.Click(`#btnCepTelDogrulamaGonder`, chromedp.ByQuery),
			chromedp.Sleep(3*time.Second),
		); err != nil {
			return false, fmt.Errorf("SMS onaylama: %w", err)
		}
		b.logger.Info("SMS doğrulama tamamlandı")
	}

	return true, nil
}

// dumpDiagnostics sonuç bulunamadığında sayfanın durumunu kaydeder.
func (b *bot) dumpDiagnostics(n int) {
	var url, bodyText string
	var shot []byte
	_ = b.run(10*time.Second,
		chromedp.Location(&url),
		chromedp.Evaluate(`document.body ? document.body.innerText.slice(0, 1500) : ''`, &bodyText),
		chromedp.CaptureScreenshot(&shot),
	)
	shotPath := fmt.Sprintf("/tmp/booking-test-attempt%d.png", n)
	if len(shot) > 0 {
		_ = os.WriteFile(shotPath, shot, 0o644)
	}
	b.logger.Warn("Tanılama",
		zap.String("url", url),
		zap.String("screenshot", shotPath),
		zap.String("body", bodyText),
	)
}

// recover hata sonrası sayfayı baştan kurar (login kontrolü dahil).
func (b *bot) recover() error {
	var url string
	_ = b.run(5*time.Second, chromedp.Location(&url))
	if strings.Contains(url, "uyegiris") {
		if err := b.login(); err != nil {
			return err
		}
	}
	return b.setupSearchPage()
}

// ── yardımcılar ──

func nextSaturday(from time.Time) time.Time {
	d := (int(time.Saturday) - int(from.Weekday()) + 7) % 7
	if d == 0 && from.Hour() >= 22 { // Cumartesi gecesi geçtiyse haftaya
		d = 7
	}
	return from.AddDate(0, 0, d)
}

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.Trim(strings.TrimSpace(v), `"'`)
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

// waitForSMSCode hem stdin'den hem proje kökündeki sms-code.txt dosyasından
// kod bekler — hangisi önce gelirse. Arka planda çalışırken stdin kapalı
// olabileceği için dosya yolu şart.
func waitForSMSCode(ctx context.Context, timeout time.Duration) string {
	const codeFile = "sms-code.txt"
	_ = os.Remove(codeFile) // eski koddan kalma dosya varsa temizle

	stdinCh := make(chan string, 1)
	go func() {
		if s := readLine(); s != "" {
			stdinCh <- s
		}
	}()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ""
		case code := <-stdinCh:
			return code
		case <-time.After(500 * time.Millisecond):
			if data, err := os.ReadFile(codeFile); err == nil {
				if code := strings.TrimSpace(string(data)); code != "" {
					_ = os.Remove(codeFile)
					return code
				}
			}
		}
	}
	return ""
}

func readLine() string {
	sc := bufio.NewScanner(os.Stdin)
	if sc.Scan() {
		return strings.TrimSpace(sc.Text())
	}
	return ""
}

func sleepCtx(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

func waitEnter(ctx context.Context) {
	done := make(chan struct{})
	go func() { readLine(); close(done) }()
	select {
	case <-ctx.Done():
	case <-done:
	}
}
