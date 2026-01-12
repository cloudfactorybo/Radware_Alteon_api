package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/service"
)

type LicenseHandler struct {
	service *service.MultiAlteonService
	logger  *logrus.Logger
}

func NewLicenseHandler(service *service.MultiAlteonService, logger *logrus.Logger) *LicenseHandler {
	return &LicenseHandler{
		service: service,
		logger:  logger,
	}
}

func (h *LicenseHandler) GetLicenses(w http.ResponseWriter, r *http.Request) {
	results, errors := h.service.GetAllLicenses()

	if len(results) == 0 {
		h.logger.Errorf("No se pudo obtener licencias de ningún Alteon")
		http.Error(w, "No se pudo obtener licencias de ningún Alteon", http.StatusInternalServerError)
		return
	}

	if len(errors) > 0 {
		h.logger.Warnf("Algunos Alteons fallaron: %d errores", len(errors))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
