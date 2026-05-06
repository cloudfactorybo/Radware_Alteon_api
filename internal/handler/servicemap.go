package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/service"
)

type ServiceMapHandler struct {
	service *service.MultiAlteonService
	logger  *logrus.Logger
}

func NewServiceMapHandler(svc *service.MultiAlteonService, logger *logrus.Logger) *ServiceMapHandler {
	return &ServiceMapHandler{service: svc, logger: logger}
}

func (h *ServiceMapHandler) GetServiceMap(w http.ResponseWriter, r *http.Request) {
	results, errs := h.service.GetAllServiceMaps(r.Context())
	writeAggregated(w, results, errs, len(results) == 0)
}
