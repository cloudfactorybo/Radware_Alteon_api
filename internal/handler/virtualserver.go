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

// GetVirtualServers devuelve los virtual servers SIN el detalle de real servers
// (más liviano: no consulta stats ni info del real server por cada servicio).
func (h *VirtualServerHandler) GetVirtualServers(w http.ResponseWriter, r *http.Request) {
	indexes := parseVServerFilter(r)
	results, errs := h.service.GetAllVirtualServers(r.Context(), indexes, false)
	writeAggregated(w, results, errs, len(results) == 0)
}

// GetRealServers devuelve lo mismo que GetVirtualServers pero CON el detalle
// completo de real servers (real_server en cada servicio). Es más pesado.
func (h *VirtualServerHandler) GetRealServers(w http.ResponseWriter, r *http.Request) {
	indexes := parseVServerFilter(r)
	results, errs := h.service.GetAllVirtualServers(r.Context(), indexes, true)
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
