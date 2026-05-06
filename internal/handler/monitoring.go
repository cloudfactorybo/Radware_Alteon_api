package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/service"
)

type MonitoringHandler struct {
	service *service.MultiAlteonService
	logger  *logrus.Logger
}

func NewMonitoringHandler(svc *service.MultiAlteonService, logger *logrus.Logger) *MonitoringHandler {
	return &MonitoringHandler{service: svc, logger: logger}
}

func (h *MonitoringHandler) GetMonitoring(w http.ResponseWriter, r *http.Request) {
	results, errs := h.service.GetAllMonitoring(r.Context())
	writeAggregated(w, results, errs, len(results) == 0)
}
