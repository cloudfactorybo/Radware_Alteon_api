package service

import (
	"context"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/cache"
	"alteon-api/internal/models"
	"alteon-api/internal/reqctx"
	"alteon-api/internal/storage"
	"alteon-api/pkg/httpclient"
)

type MultiAlteonService struct {
	repo       *storage.AlteonsRepo
	httpClient *http.Client
	logger     *logrus.Logger
	cache      *cache.Cache

	mu       sync.RWMutex
	services []*AlteonServiceWrapper
}

type AlteonServiceWrapper struct {
	Name    string
	URL     string
	IP      string
	Service *AlteonService
}

type AlteonError struct {
	Alteon string `json:"alteon"`
	Error  string `json:"error"`
}

type PingResult struct {
	Alteon string `json:"alteon"`
	OK     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
}

func NewMultiAlteonService(repo *storage.AlteonsRepo, httpClient *httpclient.Client, logger *logrus.Logger, c *cache.Cache) *MultiAlteonService {
	return &MultiAlteonService{
		repo:       repo,
		httpClient: httpClient.Client,
		logger:     logger,
		cache:      c,
	}
}

// Refresh recarga la lista de alteons desde la DB. Seguro para llamarse en caliente.
func (m *MultiAlteonService) Refresh(ctx context.Context) error {
	alteons, err := m.repo.List(ctx)
	if err != nil {
		return err
	}

	wrappers := make([]*AlteonServiceWrapper, 0, len(alteons))
	for _, a := range alteons {
		if !a.Enabled {
			continue
		}
		svc := NewAlteonService(a.Name, a.BaseURL, a.Username, a.Password, m.httpClient, m.logger, m.cache)
		wrappers = append(wrappers, &AlteonServiceWrapper{
			Name:    a.Name,
			URL:     a.BaseURL,
			IP:      extractIPFromURL(a.BaseURL),
			Service: svc,
		})
	}

	m.mu.Lock()
	prev := len(m.services)
	m.services = wrappers
	m.mu.Unlock()

	now := len(wrappers)
	if prev != now {
		m.logger.WithFields(logrus.Fields{
			"count": now,
			"prev":  prev,
		}).Info("alteons recargados — lista cambió")
	} else {
		m.logger.WithField("count", now).Debug("alteons recargados (sin cambios)")
	}
	return nil
}

func (m *MultiAlteonService) snapshot() []*AlteonServiceWrapper {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*AlteonServiceWrapper, len(m.services))
	copy(out, m.services)
	return out
}

func (m *MultiAlteonService) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.services)
}

func (m *MultiAlteonService) logError(ctx context.Context, endpoint string, w *AlteonServiceWrapper, err error) {
	fields := logrus.Fields{
		"alteon":   w.Name,
		"endpoint": endpoint,
	}
	if id := reqctx.ID(ctx); id != "" {
		fields["req_id"] = id
	}
	m.logger.WithFields(fields).WithError(err).Error("alteon call failed")
}

func (m *MultiAlteonService) PingAll(ctx context.Context) []PingResult {
	services := m.snapshot()
	out := make([]PingResult, len(services))

	var wg sync.WaitGroup
	for i, w := range services {
		wg.Add(1)
		go func(idx int, w *AlteonServiceWrapper) {
			defer wg.Done()
			if err := w.Service.Ping(ctx); err != nil {
				out[idx] = PingResult{Alteon: w.Name, OK: false, Error: err.Error()}
				return
			}
			out[idx] = PingResult{Alteon: w.Name, OK: true}
		}(i, w)
	}
	wg.Wait()
	return out
}

func (m *MultiAlteonService) GetAllSystemInfo(ctx context.Context) ([]models.SystemInfoResponse, []AlteonError) {
	services := m.snapshot()
	results := make([]*models.SystemInfoResponse, len(services))
	errs := make([]*AlteonError, len(services))

	var wg sync.WaitGroup
	for i, w := range services {
		wg.Add(1)
		go func(idx int, w *AlteonServiceWrapper) {
			defer wg.Done()

			data, err := w.Service.GetSystemInfo(ctx)
			if err != nil {
				m.logError(ctx, "system", w, err)
				errs[idx] = &AlteonError{Alteon: w.Name, Error: err.Error()}
				return
			}

			results[idx] = &models.SystemInfoResponse{
				AlteonName:                      w.Name,
				AlteonURL:                       w.URL,
				AlteonIP:                        w.IP,
				SysName:                         data.SysName,
				AgRtcTime:                       data.AgRtcTime,
				AgRtcDate:                       data.AgRtcDate,
				MpMemStatsFree:                  data.MpMemStatsFree,
				MpMemStatsTotal:                 data.MpMemStatsTotal,
				AgSwitchLastApplyTime:           data.AgSwitchLastApplyTime,
				AgSwitchLastSaveTime:            data.AgSwitchLastSaveTime,
				AgSwitchLastBootTime:            data.AgSwitchLastBootTime,
				AgSwitchUpTime:                  data.AgSwitchUpTime,
				AgFipsSecurityLevel:             data.AgFipsSecurityLevel,
				AgFipsNonApprovedMode:           data.AgFipsNonApprovedMode,
				MgmtPortInfoIPv6SLAACTot:        data.MgmtPortInfoIPv6SLAACTot,
				AgMgmtCurCfgIpAddr:              data.AgMgmtCurCfgIpAddr,
				AgMgmtCurCfgMask:                data.AgMgmtCurCfgMask,
				AgMgmtCurCfgGateway:             data.AgMgmtCurCfgGateway,
				AgMgmtCurCfgIpv6Addr:            data.AgMgmtCurCfgIpv6Addr,
				AgMgmtCurCfgIpv6PrefixLen:       data.AgMgmtCurCfgIpv6PrefixLen,
				AgMgmtCurCfgIpv6Gateway:         data.AgMgmtCurCfgIpv6Gateway,
				MgmtPortInfoIPv6SLAAC1Addr:      data.MgmtPortInfoIPv6SLAAC1Addr,
				MgmtPortInfoIPv6SLAAC1PrefixLen: data.MgmtPortInfoIPv6SLAAC1PrefixLen,
				MgmtPortInfoIPv6SLAAC2Addr:      data.MgmtPortInfoIPv6SLAAC2Addr,
				MgmtPortInfoIPv6SLAAC2PrefixLen: data.MgmtPortInfoIPv6SLAAC2PrefixLen,
				MgmtPortInfoIPv6SLAAC3Addr:      data.MgmtPortInfoIPv6SLAAC3Addr,
				MgmtPortInfoIPv6SLAAC3PrefixLen: data.MgmtPortInfoIPv6SLAAC3PrefixLen,
				MgmtPortInfoIPv6SLAAC4Addr:      data.MgmtPortInfoIPv6SLAAC4Addr,
				MgmtPortInfoIPv6SLAAC4PrefixLen: data.MgmtPortInfoIPv6SLAAC4PrefixLen,
				HwMACAddress:                    data.HwMACAddress,
			}
		}(i, w)
	}
	wg.Wait()

	return collectSystem(results), collectErrors(errs)
}

func (m *MultiAlteonService) GetAllLicenses(ctx context.Context) ([]models.LicenseResponseWrapper, []AlteonError) {
	services := m.snapshot()
	results := make([]*models.LicenseResponseWrapper, len(services))
	errs := make([]*AlteonError, len(services))

	var wg sync.WaitGroup
	for i, w := range services {
		wg.Add(1)
		go func(idx int, w *AlteonServiceWrapper) {
			defer wg.Done()
			data, err := w.Service.GetLicenses(ctx)
			if err != nil {
				m.logError(ctx, "licenses", w, err)
				errs[idx] = &AlteonError{Alteon: w.Name, Error: err.Error()}
				return
			}
			results[idx] = &models.LicenseResponseWrapper{
				AlteonName: w.Name,
				AlteonURL:  w.URL,
				AlteonIP:   w.IP,
				Licenses:   data.Licenses,
			}
		}(i, w)
	}
	wg.Wait()

	return collectLicenses(results), collectErrors(errs)
}

func (m *MultiAlteonService) GetAllVirtualServers(ctx context.Context) ([]models.VirtualServersResponseWrapper, []AlteonError) {
	services := m.snapshot()
	results := make([]*models.VirtualServersResponseWrapper, len(services))
	errs := make([]*AlteonError, len(services))

	var wg sync.WaitGroup
	for i, w := range services {
		wg.Add(1)
		go func(idx int, w *AlteonServiceWrapper) {
			defer wg.Done()
			data, err := w.Service.GetVirtualServers(ctx)
			if err != nil {
				m.logError(ctx, "virtualservers", w, err)
				errs[idx] = &AlteonError{Alteon: w.Name, Error: err.Error()}
				return
			}
			results[idx] = &models.VirtualServersResponseWrapper{
				AlteonName:     w.Name,
				AlteonURL:      w.URL,
				AlteonIP:       w.IP,
				VirtualServers: data.VirtualServers,
			}
		}(i, w)
	}
	wg.Wait()

	return collectVirtualServers(results), collectErrors(errs)
}

func (m *MultiAlteonService) GetAllMonitoring(ctx context.Context) ([]models.MonitoringResponseWrapper, []AlteonError) {
	services := m.snapshot()
	results := make([]*models.MonitoringResponseWrapper, len(services))
	errs := make([]*AlteonError, len(services))

	var wg sync.WaitGroup
	for i, w := range services {
		wg.Add(1)
		go func(idx int, w *AlteonServiceWrapper) {
			defer wg.Done()
			data, err := w.Service.GetMonitoring(ctx)
			if err != nil {
				m.logError(ctx, "monitoring", w, err)
				errs[idx] = &AlteonError{Alteon: w.Name, Error: err.Error()}
				return
			}
			results[idx] = &models.MonitoringResponseWrapper{
				AlteonName: w.Name,
				AlteonURL:  w.URL,
				AlteonIP:   w.IP,
				CPU:        data.CPU,
				Memory:     data.Memory,
				Cores:      data.Cores,
			}
		}(i, w)
	}
	wg.Wait()

	return collectMonitoring(results), collectErrors(errs)
}

func (m *MultiAlteonService) GetAllServiceMaps(ctx context.Context) ([]models.ServiceMapResponseWrapper, []AlteonError) {
	services := m.snapshot()
	results := make([]*models.ServiceMapResponseWrapper, len(services))
	errs := make([]*AlteonError, len(services))

	var wg sync.WaitGroup
	for i, w := range services {
		wg.Add(1)
		go func(idx int, w *AlteonServiceWrapper) {
			defer wg.Done()
			data, err := w.Service.GetServiceMap(ctx)
			if err != nil {
				m.logError(ctx, "servicemap", w, err)
				errs[idx] = &AlteonError{Alteon: w.Name, Error: err.Error()}
				return
			}
			results[idx] = &models.ServiceMapResponseWrapper{
				AlteonName: w.Name,
				AlteonURL:  w.URL,
				AlteonIP:   w.IP,
				Timestamp:  data.Timestamp,
				VServers:   data.VServers,
				Status:     data.Status,
			}
		}(i, w)
	}
	wg.Wait()

	return collectServiceMaps(results), collectErrors(errs)
}

func collectErrors(src []*AlteonError) []AlteonError {
	out := []AlteonError{}
	for _, e := range src {
		if e != nil {
			out = append(out, *e)
		}
	}
	return out
}

func collectSystem(src []*models.SystemInfoResponse) []models.SystemInfoResponse {
	out := []models.SystemInfoResponse{}
	for _, r := range src {
		if r != nil {
			out = append(out, *r)
		}
	}
	return out
}

func collectLicenses(src []*models.LicenseResponseWrapper) []models.LicenseResponseWrapper {
	out := []models.LicenseResponseWrapper{}
	for _, r := range src {
		if r != nil {
			out = append(out, *r)
		}
	}
	return out
}

func collectVirtualServers(src []*models.VirtualServersResponseWrapper) []models.VirtualServersResponseWrapper {
	out := []models.VirtualServersResponseWrapper{}
	for _, r := range src {
		if r != nil {
			out = append(out, *r)
		}
	}
	return out
}

func collectMonitoring(src []*models.MonitoringResponseWrapper) []models.MonitoringResponseWrapper {
	out := []models.MonitoringResponseWrapper{}
	for _, r := range src {
		if r != nil {
			out = append(out, *r)
		}
	}
	return out
}

func collectServiceMaps(src []*models.ServiceMapResponseWrapper) []models.ServiceMapResponseWrapper {
	out := []models.ServiceMapResponseWrapper{}
	for _, r := range src {
		if r != nil {
			out = append(out, *r)
		}
	}
	return out
}
