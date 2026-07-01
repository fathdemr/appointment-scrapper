package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"appointment-scrapper/config"
	"appointment-scrapper/internal/scheduler"
)

type Server struct {
	cfg       config.AppConfig
	scheduler *scheduler.Scheduler
	logger    *zap.Logger
	engine    *gin.Engine
}

func New(cfg config.AppConfig, sc *scheduler.Scheduler, logger *zap.Logger) *Server {
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		cfg:       cfg,
		scheduler: sc,
		logger:    logger,
		engine:    gin.New(),
	}

	s.engine.Use(gin.Recovery())
	s.registerRoutes()
	return s
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.logger.Info("API sunucusu başlatılıyor", zap.String("addr", addr))
	return s.engine.Run(addr)
}

func (s *Server) registerRoutes() {
	s.engine.GET("/health", s.health)
	s.engine.GET("/status", s.status)
	s.engine.POST("/run", s.runNow)
	s.engine.POST("/start", s.start)
	s.engine.POST("/stop", s.stop)
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) status(c *gin.Context) {
	c.JSON(http.StatusOK, s.scheduler.GetStatus())
}

func (s *Server) runNow(c *gin.Context) {
	s.scheduler.RunNow()
	c.JSON(http.StatusAccepted, gin.H{"message": "Scraper tetiklendi"})
}

func (s *Server) start(c *gin.Context) {
	s.scheduler.Start()
	c.JSON(http.StatusOK, gin.H{"message": "Zamanlayıcı başlatıldı"})
}

func (s *Server) stop(c *gin.Context) {
	s.scheduler.Stop()
	c.JSON(http.StatusOK, gin.H{"message": "Zamanlayıcı durduruldu"})
}
