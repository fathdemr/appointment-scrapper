package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"appointment-scrapper/config"
	"appointment-scrapper/internal/api/handler"
	catalogrepo "appointment-scrapper/repository/catalog"
	"appointment-scrapper/service"
)

type Server struct {
	cfg         config.AppConfig
	svc         service.JobService
	catalogRepo *catalogrepo.CatalogRepository
	logger      *zap.Logger
	engine      *gin.Engine
}

func New(cfg config.AppConfig, svc service.JobService, catalogRepo *catalogrepo.CatalogRepository, logger *zap.Logger) *Server {
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		cfg:         cfg,
		svc:         svc,
		catalogRepo: catalogRepo,
		logger:      logger,
		engine:      gin.New(),
	}

	s.engine.Use(gin.Recovery())
	s.engine.Use(corsMiddleware())
	s.registerRoutes()
	return s
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.logger.Info("API sunucusu başlatılıyor", zap.String("addr", addr))
	return s.engine.Run(addr)
}

func (s *Server) registerRoutes() {
	jobH := handler.NewJobHandler(s.svc, s.logger)
	sysH := handler.NewSystemHandler(s.logger)
	catalogH := handler.NewCatalogHandler(s.catalogRepo, s.logger)

	// Auth gerektirmeyen endpoint
	s.engine.GET("/health", sysH.Health)

	v1 := s.engine.Group("/api/v1")
	if s.cfg.APIKey != "" {
		v1.Use(apiKeyMiddleware(s.cfg.APIKey))
	}

	v1.GET("/system/status", sysH.SystemStatus)

	jobs := v1.Group("/jobs")
	jobs.POST("", jobH.Create)
	jobs.GET("", jobH.List)
	jobs.GET("/:id", jobH.Get)
	jobs.PUT("/:id", jobH.Update)
	jobs.DELETE("/:id", jobH.Delete)
	jobs.POST("/:id/start", jobH.Start)
	jobs.POST("/:id/stop", jobH.Stop)
	jobs.POST("/:id/run", jobH.RunNow)
	jobs.GET("/:id/status", jobH.Status)
	jobs.POST("/:id/telegram/verify", jobH.VerifyTelegram)
	jobs.POST("/:id/sms-reply", jobH.SubmitSMSCode)

	cat := v1.Group("/catalog")
	cat.GET("/sport-types", catalogH.GetSportTypes)
	cat.GET("/facilities", catalogH.GetFacilities)
	cat.GET("/courts", catalogH.GetCourts)

	tgSetup := handler.NewTelegramSetupHandler(s.logger)
	tg := v1.Group("/telegram")
	tg.GET("/verify-token", tgSetup.VerifyBotToken)
	tg.GET("/detect-chat", tgSetup.DetectChatID)
}

// corsMiddleware UI'ın farklı origin'den istek atabilmesi için CORS başlıklarını ekler.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// apiKeyMiddleware X-API-Key header'ı veya api_key query parametresi ile auth yapar.
func apiKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("api_key")
		}
		if key != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "geçersiz API anahtarı"})
			return
		}
		c.Next()
	}
}
