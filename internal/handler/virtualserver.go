package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/service"
)

type VirtualServerHandler struct {
	service *service.MultiAlteonService
	logger  *logrus.Logger
}

func NewVirtualServerHandler(service *service.MultiAlteonService, logger *logrus.Logger) *VirtualServerHandler {
	return &VirtualServerHandler{
		service: service,
		logger:  logger,
	}
}

func (h *VirtualServerHandler) GetVirtualServers(w http.ResponseWriter, r *http.Request) {
	results, errors := h.service.GetAllVirtualServers()

	if len(results) == 0 {
		h.logger.Errorf("No se pudo obtener virtual servers de ningún Alteon")
		http.Error(w, "No se pudo obtener virtual servers de ningún Alteon", http.StatusInternalServerError)
		return
	}

	if len(errors) > 0 {
		h.logger.Warnf("Algunos Alteons fallaron: %d errores", len(errors))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
