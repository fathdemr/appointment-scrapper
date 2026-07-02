package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"appointment-scrapper/repository"
	"appointment-scrapper/service"
)

// JobHandler booking job'ları için HTTP endpoint'lerini yönetir.
type JobHandler struct {
	svc    service.JobService
	logger *zap.Logger
}

func NewJobHandler(svc service.JobService, logger *zap.Logger) *JobHandler {
	return &JobHandler{svc: svc, logger: logger}
}

// Create godoc
// @Summary      Yeni booking job oluştur
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Param        job  body      service.CreateJobRequest  true  "Job yapılandırması"
// @Success      201  {object}  model.BookingJob
// @Failure      400  {object}  map[string]string
// @Router       /api/v1/jobs [post]
func (h *JobHandler) Create(c *gin.Context) {
	var req service.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.svc.CreateJob(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("job oluşturulamadı", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, job)
}

// List godoc
// @Summary      Tüm booking job'larını listele
// @Tags         jobs
// @Produce      json
// @Success      200  {array}   model.BookingJob
// @Router       /api/v1/jobs [get]
func (h *JobHandler) List(c *gin.Context) {
	jobs, err := h.svc.ListJobs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// Get godoc
// @Summary      Bir booking job'ı getir
// @Tags         jobs
// @Produce      json
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  model.BookingJob
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/jobs/{id} [get]
func (h *JobHandler) Get(c *gin.Context) {
	job, err := h.svc.GetJob(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "job bulunamadı"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

// Update godoc
// @Summary      Booking job'ı güncelle
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Param        id   path      string                    true  "Job ID"
// @Param        job  body      service.UpdateJobRequest  true  "Güncellenen yapılandırma"
// @Success      200  {object}  model.BookingJob
// @Failure      400  {object}  map[string]string
// @Router       /api/v1/jobs/{id} [put]
func (h *JobHandler) Update(c *gin.Context) {
	var req service.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.svc.UpdateJob(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "job bulunamadı"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

// Delete godoc
// @Summary      Booking job'ı sil
// @Tags         jobs
// @Param        id   path  string  true  "Job ID"
// @Success      204
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/jobs/{id} [delete]
func (h *JobHandler) Delete(c *gin.Context) {
	if err := h.svc.DeleteJob(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "job bulunamadı"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Start godoc
// @Summary      Job zamanlayıcısını başlat
// @Tags         jobs
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/jobs/{id}/start [post]
func (h *JobHandler) Start(c *gin.Context) {
	if err := h.svc.StartJob(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "job bulunamadı"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "zamanlayıcı başlatıldı"})
}

// Stop godoc
// @Summary      Job zamanlayıcısını durdur
// @Tags         jobs
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/jobs/{id}/stop [post]
func (h *JobHandler) Stop(c *gin.Context) {
	if err := h.svc.StopJob(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "zamanlayıcı durduruldu"})
}

// RunNow godoc
// @Summary      Job'ı hemen çalıştır (manuel tetikleme)
// @Tags         jobs
// @Param        id   path      string  true  "Job ID"
// @Success      202  {object}  map[string]string
// @Router       /api/v1/jobs/{id}/run [post]
func (h *JobHandler) RunNow(c *gin.Context) {
	if err := h.svc.RunJobNow(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "job bulunamadı"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "scraper tetiklendi"})
}

// Status godoc
// @Summary      Job çalışma durumunu getir (DB + anlık zamanlayıcı)
// @Tags         jobs
// @Produce      json
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  service.JobStatusResponse
// @Router       /api/v1/jobs/{id}/status [get]
func (h *JobHandler) Status(c *gin.Context) {
	status, err := h.svc.GetJobStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "job bulunamadı"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// VerifyTelegram godoc
// @Summary      Telegram bot bağlantısını test et
// @Tags         jobs
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/jobs/{id}/telegram/verify [post]
func (h *JobHandler) VerifyTelegram(c *gin.Context) {
	if err := h.svc.VerifyTelegram(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Telegram mesajı gönderildi"})
}

// SubmitSMSCode godoc
// @Summary      SMS doğrulama kodunu ilet (UI üzerinden)
// @Tags         jobs
// @Accept       json
// @Param        id    path      string  true  "Job ID"
// @Param        body  body      object  true  "SMS kodu"
// @Success      200   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Router       /api/v1/jobs/{id}/sms-reply [post]
func (h *JobHandler) SubmitSMSCode(c *gin.Context) {
	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.svc.SubmitSMSCode(c.Param("id"), body.Code) {
		c.JSON(http.StatusConflict, gin.H{"error": "bu job için SMS kodu beklenmiyor"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SMS kodu iletildi"})
}
