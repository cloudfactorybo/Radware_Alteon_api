package service

import (
	"sync"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/config"
	"alteon-api/internal/models"
	"alteon-api/pkg/httpclient"
)

type MultiAlteonService struct {
	services []*AlteonServiceWrapper
	logger   *logrus.Logger
}

type AlteonServiceWrapper struct {
	Name    string
	URL     string
	IP      string
	Service *AlteonService
}

func NewMultiAlteonService(configs []config.AlteonConfig, httpClient *httpclient.Client, logger *logrus.Logger) *MultiAlteonService {
	services := make([]*AlteonServiceWrapper, 0, len(configs))

	for _, cfg := range configs {
		service := NewAlteonService(
			cfg.BaseURL,
			cfg.Username,
			cfg.Password,
			httpClient.Client,
			logger,
		)

		services = append(services, &AlteonServiceWrapper{
			Name:    cfg.Name,
			URL:     cfg.BaseURL,
			IP:      extractIPFromURL(cfg.BaseURL),
			Service: service,
		})

		logger.Infof("Configurado Alteon: %s (%s)", cfg.Name, cfg.BaseURL)
	}

	return &MultiAlteonService{
		services: services,
		logger:   logger,
	}
}

func (m *MultiAlteonService) GetAllSystemInfo() ([]models.SystemInfoResponse, []error) {
	var wg sync.WaitGroup
	results := make([]models.SystemInfoResponse, len(m.services))
	errors := make([]error, len(m.services))

	for i, wrapper := range m.services {
		wg.Add(1)
		go func(index int, w *AlteonServiceWrapper) {
			defer wg.Done()

			data, err := w.Service.GetSystemInfo()
			if err != nil {
				m.logger.Errorf("Error obteniendo system info de %s: %v", w.Name, err)
				errors[index] = err
				return
			}

			results[index] = models.SystemInfoResponse{
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
		}(i, wrapper)
	}

	wg.Wait()

	// Filtrar resultados exitosos
	successResults := []models.SystemInfoResponse{}
	successErrors := []error{}
	for i, result := range results {
		if errors[i] == nil {
			successResults = append(successResults, result)
		} else {
			successErrors = append(successErrors, errors[i])
		}
	}

	return successResults, successErrors
}

func (m *MultiAlteonService) GetAllLicenses() ([]models.LicenseResponseWrapper, []error) {
	var wg sync.WaitGroup
	results := make([]models.LicenseResponseWrapper, len(m.services))
	errors := make([]error, len(m.services))

	for i, wrapper := range m.services {
		wg.Add(1)
		go func(index int, w *AlteonServiceWrapper) {
			defer wg.Done()

			data, err := w.Service.GetLicenses()
			if err != nil {
				m.logger.Errorf("Error obteniendo licenses de %s: %v", w.Name, err)
				errors[index] = err
				return
			}

			results[index] = models.LicenseResponseWrapper{
				AlteonName: w.Name,
				AlteonURL:  w.URL,
				AlteonIP:   w.IP,
				Licenses:   data.Licenses,
			}
		}(i, wrapper)
	}

	wg.Wait()

	// Filtrar resultados exitosos
	successResults := []models.LicenseResponseWrapper{}
	successErrors := []error{}
	for i, result := range results {
		if errors[i] == nil {
			successResults = append(successResults, result)
		} else {
			successErrors = append(successErrors, errors[i])
		}
	}

	return successResults, successErrors
}

func (m *MultiAlteonService) GetAllVirtualServers() ([]models.VirtualServersResponseWrapper, []error) {
	var wg sync.WaitGroup
	results := make([]models.VirtualServersResponseWrapper, len(m.services))
	errors := make([]error, len(m.services))

	for i, wrapper := range m.services {
		wg.Add(1)
		go func(index int, w *AlteonServiceWrapper) {
			defer wg.Done()

			data, err := w.Service.GetVirtualServers()
			if err != nil {
				m.logger.Errorf("Error obteniendo virtual servers de %s: %v", w.Name, err)
				errors[index] = err
				return
			}

			results[index] = models.VirtualServersResponseWrapper{
				AlteonName:     w.Name,
				AlteonURL:      w.URL,
				AlteonIP:       w.IP,
				VirtualServers: data.VirtualServers,
			}
		}(i, wrapper)
	}

	wg.Wait()

	// Filtrar resultados exitosos
	successResults := []models.VirtualServersResponseWrapper{}
	successErrors := []error{}
	for i, result := range results {
		if errors[i] == nil {
			successResults = append(successResults, result)
		} else {
			successErrors = append(successErrors, errors[i])
		}
	}

	return successResults, successErrors
}

func (m *MultiAlteonService) GetAllMonitoring() ([]models.MonitoringResponseWrapper, []error) {
	var wg sync.WaitGroup
	results := make([]models.MonitoringResponseWrapper, len(m.services))
	errors := make([]error, len(m.services))

	for i, wrapper := range m.services {
		wg.Add(1)
		go func(index int, w *AlteonServiceWrapper) {
			defer wg.Done()

			data, err := w.Service.GetMonitoring()
			if err != nil {
				m.logger.Errorf("Error obteniendo monitoring de %s: %v", w.Name, err)
				errors[index] = err
				return
			}

			results[index] = models.MonitoringResponseWrapper{
				AlteonName: w.Name,
				AlteonURL:  w.URL,
				AlteonIP:   w.IP,
				CPU:        data.CPU,
				Memory:     data.Memory,
				Cores:      data.Cores,
			}
		}(i, wrapper)
	}

	wg.Wait()

	// Filtrar resultados exitosos
	successResults := []models.MonitoringResponseWrapper{}
	successErrors := []error{}
	for i, result := range results {
		if errors[i] == nil {
			successResults = append(successResults, result)
		} else {
			successErrors = append(successErrors, errors[i])
		}
	}

	return successResults, successErrors
}

func (m *MultiAlteonService) GetAllServiceMaps() ([]models.ServiceMapResponseWrapper, []error) {
	var wg sync.WaitGroup
	results := make([]models.ServiceMapResponseWrapper, len(m.services))
	errors := make([]error, len(m.services))

	for i, wrapper := range m.services {
		wg.Add(1)
		go func(index int, w *AlteonServiceWrapper) {
			defer wg.Done()

			data, err := w.Service.GetServiceMap()
			if err != nil {
				m.logger.Errorf("Error obteniendo service map de %s: %v", w.Name, err)
				errors[index] = err
				return
			}

			results[index] = models.ServiceMapResponseWrapper{
				AlteonName: w.Name,
				AlteonURL:  w.URL,
				AlteonIP:   w.IP,
				Timestamp:  data.Timestamp,
				VServers:   data.VServers,
				Status:     data.Status,
			}
		}(i, wrapper)
	}

	wg.Wait()

	// Filtrar resultados exitosos
	successResults := []models.ServiceMapResponseWrapper{}
	successErrors := []error{}
	for i, result := range results {
		if errors[i] == nil {
			successResults = append(successResults, result)
		} else {
			successErrors = append(successErrors, errors[i])
		}
	}

	return successResults, successErrors
}
