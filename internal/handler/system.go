package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/service"
)

type SystemHandler struct {
	service *service.MultiAlteonService
	logger  *logrus.Logger
}

func NewSystemHandler(svc *service.MultiAlteonService, logger *logrus.Logger) *SystemHandler {
	return &SystemHandler{service: svc, logger: logger}
}

func (h *SystemHandler) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	results, errs := h.service.GetAllSystemInfo(r.Context())
	writeAggregated(w, results, errs, len(results) == 0)
}
