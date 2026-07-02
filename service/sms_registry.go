package service

import "sync"

// SMSCodeRegistry SMS doğrulama kodlarını scraper goroutine'ine iletmek için
// in-memory bir kanal haritası tutar. UI veya Telegram'dan gelen kod buraya yazılır.
type SMSCodeRegistry struct {
	mu       sync.RWMutex
	channels map[string]chan string // jobID → channel
}

func NewSMSCodeRegistry() *SMSCodeRegistry {
	return &SMSCodeRegistry{
		channels: make(map[string]chan string),
	}
}

// Register bir job için SMS kod kanalı oluşturur ve döner.
// Scraper bu kanaldan okur.
func (r *SMSCodeRegistry) Register(jobID string) chan string {
	ch := make(chan string, 1)
	r.mu.Lock()
	r.channels[jobID] = ch
	r.mu.Unlock()
	return ch
}

// Submit API veya UI'dan gelen SMS kodunu ilgili scraper kanalına yazar.
// Job beklemiyorsa false döner.
func (r *SMSCodeRegistry) Submit(jobID string, code string) bool {
	r.mu.RLock()
	ch, ok := r.channels[jobID]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	select {
	case ch <- code:
		return true
	default:
		return false
	}
}

// Unregister kanalı temizler (WaitForReply bittikten sonra çağrılır).
func (r *SMSCodeRegistry) Unregister(jobID string) {
	r.mu.Lock()
	delete(r.channels, jobID)
	r.mu.Unlock()
}

// IsWaiting job'ın şu an SMS kodu bekleyip beklemediğini döner.
func (r *SMSCodeRegistry) IsWaiting(jobID string) bool {
	r.mu.RLock()
	_, ok := r.channels[jobID]
	r.mu.RUnlock()
	return ok
}
