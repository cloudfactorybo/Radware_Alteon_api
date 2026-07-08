package handler

import (
	"net/http"
	"strings"

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
	indexes := parseVServerFilter(r)
	results, errs := h.service.GetAllVirtualServers(r.Context(), indexes)
	writeAggregated(w, results, errs, len(results) == 0)
}

// parseVServerFilter lee el/los parámetros ?index=... para consultar solo ciertos
// virtual servers. Admite CSV y repetidos: ?index=1,2 o ?index=1&index=2.
// Lista vacía = todos los virtual servers (comportamiento por defecto).
func parseVServerFilter(r *http.Request) []string {
	var out []string
	for _, v := range r.URL.Query()["index"] {
		for _, part := range strings.Split(v, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}
