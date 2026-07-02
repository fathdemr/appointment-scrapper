package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	catalogrepo "appointment-scrapper/repository/catalog"
)

type CatalogHandler struct {
	repo   *catalogrepo.CatalogRepository
	logger *zap.Logger
}

func NewCatalogHandler(repo *catalogrepo.CatalogRepository, logger *zap.Logger) *CatalogHandler {
	return &CatalogHandler{repo: repo, logger: logger}
}

// GetSportTypes godoc
// @Summary  Tüm spor dallarını listeler
// @Tags     catalog
// @Produce  json
// @Success  200 {array} model.SportType
// @Router   /catalog/sport-types [get]
func (h *CatalogHandler) GetSportTypes(c *gin.Context) {
	items, err := h.repo.ListSportTypes(c.Request.Context())
	if err != nil {
		h.logger.Error("sport types listelenemedi", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetFacilities godoc
// @Summary  Spor tipine göre tesis listesi
// @Tags     catalog
// @Produce  json
// @Param    sport_type_id query string true "SportType ID"
// @Success  200 {array} model.Facility
// @Router   /catalog/facilities [get]
func (h *CatalogHandler) GetFacilities(c *gin.Context) {
	sportTypeID := c.Query("sport_type_id")
	if sportTypeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sport_type_id gerekli"})
		return
	}
	items, err := h.repo.ListFacilities(c.Request.Context(), sportTypeID)
	if err != nil {
		h.logger.Error("facilities listelenemedi", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetCourts godoc
// @Summary  Tesise göre salon/kort listesi
// @Tags     catalog
// @Produce  json
// @Param    facility_id query string true "Facility ID"
// @Success  200 {array} model.Court
// @Router   /catalog/courts [get]
func (h *CatalogHandler) GetCourts(c *gin.Context) {
	facilityID := c.Query("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "facility_id gerekli"})
		return
	}
	items, err := h.repo.ListCourts(c.Request.Context(), facilityID)
	if err != nil {
		h.logger.Error("courts listelenemedi", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}
