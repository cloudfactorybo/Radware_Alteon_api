package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/service"
)

type GatewayHandler struct {
	service *service.MultiAlteonService
	logger  *logrus.Logger
}

func NewGatewayHandler(svc *service.MultiAlteonService, logger *logrus.Logger) *GatewayHandler {
	return &GatewayHandler{service: svc, logger: logger}
}

func (h *GatewayHandler) GetGateways(w http.ResponseWriter, r *http.Request) {
	results, errs := h.service.GetAllGateways(r.Context())
	writeAggregated(w, results, errs, len(results) == 0)
}
