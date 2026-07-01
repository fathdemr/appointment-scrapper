package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"appointment-scrapper/config"
)

type TelegramNotifier struct {
	botToken string
	chatID   string
}

func NewTelegram(cfg config.TelegramConfig) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: cfg.BotToken,
		chatID:   cfg.ChatID,
	}
}

func (t *TelegramNotifier) Send(ctx context.Context, msg string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id":    {t.chatID},
		"text":       {msg},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}
	return nil
}

// WaitForReply Telegram'a prompt gönderir, ardından kullanıcının
// gönderdiği ilk mesajı (SMS kodu) döner. timeout süresi kadar bekler.
func (t *TelegramNotifier) WaitForReply(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	// Önce mevcut update offset'i al (eski mesajları atla)
	offset, err := t.getCurrentOffset()
	if err != nil {
		offset = 0
	}

	// Kullanıcıya bildirim gönder
	if err := t.Send(ctx, prompt); err != nil {
		return "", fmt.Errorf("prompt gönderilemedi: %w", err)
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		remaining := time.Until(deadline)
		waitSecs := int(remaining.Seconds())
		if waitSecs > 25 {
			waitSecs = 25 // Telegram long-poll max 25 sn
		}
		if waitSecs <= 0 {
			break
		}

		text, newOffset, err := t.pollUpdates(offset, waitSecs)
		if err == nil && newOffset > offset {
			offset = newOffset
		}
		if text != "" {
			return text, nil
		}
	}

	return "", fmt.Errorf("timeout: %v içinde yanıt gelmedi", timeout)
}

type tgUpdate struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

type tgUpdatesResp struct {
	OK     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

func (t *TelegramNotifier) getCurrentOffset() (int, error) {
	// -1 timeout = sadece birikmiş mesajları al
	_, offset, err := t.pollUpdates(0, 0)
	return offset, err
}

func (t *TelegramNotifier) pollUpdates(offset, timeoutSecs int) (text string, newOffset int, err error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", t.botToken)
	params := url.Values{
		"offset":  {strconv.Itoa(offset)},
		"timeout": {strconv.Itoa(timeoutSecs)},
		"limit":   {"10"},
	}

	client := &http.Client{Timeout: time.Duration(timeoutSecs+5) * time.Second}
	resp, err := client.PostForm(apiURL, params)
	if err != nil {
		return "", offset, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result tgUpdatesResp
	if err := json.Unmarshal(body, &result); err != nil {
		return "", offset, err
	}

	chatID, _ := strconv.ParseInt(t.chatID, 10, 64)
	newOffset = offset

	for _, u := range result.Result {
		newOffset = u.UpdateID + 1
		// Sadece bizim chat'ten gelen mesajları dikkate al
		if u.Message.Chat.ID == chatID && u.Message.Text != "" {
			return u.Message.Text, newOffset, nil
		}
	}

	return "", newOffset, nil
}

func FormatBookingMessage(r BookingResult, extra string) string {
	msg := fmt.Sprintf(
		"✅ <b>Randevu Sepete Eklendi!</b>\n\n"+
			"🏟 <b>Tesis:</b> %s\n"+
			"🏃 <b>Spor:</b> %s\n"+
			"🎯 <b>Alan:</b> %s\n"+
			"📅 <b>Tarih:</b> %s\n"+
			"🕐 <b>Saat:</b> %s\n\n"+
			"⚠️ %s",
		r.Facility, r.SportType, r.Court, r.Date, r.Time, extra,
	)
	if r.CartURL != "" {
		msg += fmt.Sprintf("\n\n🔗 <a href=\"%s\">Sepete Git</a>", r.CartURL)
	}
	return msg
}
