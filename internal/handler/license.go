package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/service"
)

type LicenseHandler struct {
	service *service.MultiAlteonService
	logger  *logrus.Logger
}

func NewLicenseHandler(svc *service.MultiAlteonService, logger *logrus.Logger) *LicenseHandler {
	return &LicenseHandler{service: svc, logger: logger}
}

func (h *LicenseHandler) GetLicenses(w http.ResponseWriter, r *http.Request) {
	results, errs := h.service.GetAllLicenses(r.Context())
	writeAggregated(w, results, errs, len(results) == 0)
}
