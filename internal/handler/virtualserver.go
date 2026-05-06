package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/service"
)

type VirtualServerHandler struct {
	service *service.MultiAlteonService
	logger  *logrus.Logger
}

func NewVirtualServerHandler(svc *service.MultiAlteonService, logger *logrus.Logger) *VirtualServerHandler {
	return &VirtualServerHandler{service: svc, logger: logger}
}

func (h *VirtualServerHandler) GetVirtualServers(w http.ResponseWriter, r *http.Request) {
	results, errs := h.service.GetAllVirtualServers(r.Context())
	writeAggregated(w, results, errs, len(results) == 0)
}
