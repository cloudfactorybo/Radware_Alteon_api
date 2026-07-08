package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/service"
)

type WanLinkHandler struct {
	service *service.MultiAlteonService
	logger  *logrus.Logger
}

func NewWanLinkHandler(svc *service.MultiAlteonService, logger *logrus.Logger) *WanLinkHandler {
	return &WanLinkHandler{service: svc, logger: logger}
}

func (h *WanLinkHandler) GetSmartNat(w http.ResponseWriter, r *http.Request) {
	results, errs := h.service.GetAllSmartNat(r.Context())
	writeAggregated(w, results, errs, len(results) == 0)
}

func (h *WanLinkHandler) GetWanLinkGroups(w http.ResponseWriter, r *http.Request) {
	results, errs := h.service.GetAllWanLinkGroups(r.Context())
	writeAggregated(w, results, errs, len(results) == 0)
}

func (h *WanLinkHandler) GetWanLinks(w http.ResponseWriter, r *http.Request) {
	results, errs := h.service.GetAllWanLinks(r.Context())
	writeAggregated(w, results, errs, len(results) == 0)
}
