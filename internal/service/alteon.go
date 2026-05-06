package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/cache"
	"alteon-api/internal/models"
	"alteon-api/internal/reqctx"
)

const (
	maxConcurrentRequests = 8                // por alteon
	statsTTL              = 15 * time.Second // cache para stats/realserver info
)

type AlteonService struct {
	name       string
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	logger     *logrus.Logger
	cache      *cache.Cache
	sem        chan struct{}
}

func NewAlteonService(name, baseURL, username, password string, client *http.Client, logger *logrus.Logger, c *cache.Cache) *AlteonService {
	return &AlteonService{
		name:       name,
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: client,
		logger:     logger,
		cache:      c,
		sem:        make(chan struct{}, maxConcurrentRequests),
	}
}

func (s *AlteonService) Name() string { return s.name }

func (s *AlteonService) log() *logrus.Entry {
	return s.logger.WithField("alteon", s.name)
}

// logCtx es como log() pero añade el req_id si está en el contexto.
func (s *AlteonService) logCtx(ctx context.Context) *logrus.Entry {
	entry := s.log()
	if id := reqctx.ID(ctx); id != "" {
		entry = entry.WithField("req_id", id)
	}
	return entry
}

func (s *AlteonService) makeRequest(ctx context.Context, endpoint string) ([]byte, error) {
	select {
	case s.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-s.sem }()

	url := fmt.Sprintf("%s%s", s.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creando request: %w", err)
	}

	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	s.log().WithField("url", url).Debug("request al alteon")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ejecutando request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("leyendo respuesta: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("auth 401: %s", strings.TrimSpace(string(body)))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func (s *AlteonService) makeRequestCached(ctx context.Context, endpoint string, ttl time.Duration) ([]byte, error) {
	if ttl <= 0 || s.cache == nil {
		return s.makeRequest(ctx, endpoint)
	}
	key := s.name + ":" + endpoint
	if cached, ok := s.cache.Get(key); ok {
		return cached, nil
	}
	body, err := s.makeRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	s.cache.Set(key, body, ttl)
	return body, nil
}

func (s *AlteonService) Ping(ctx context.Context) error {
	_, err := s.makeRequest(ctx, "/config?prop=sysName")
	return err
}

func (s *AlteonService) GetSystemInfo(ctx context.Context) (*models.SystemInfo, error) {
	endpoint := "/config?prop=sysName,agRtcTime,mpMemStatsFree,agRtcDate,mpMemStatsTotal,agSwitchLastApplyTime,agSwitchLastSaveTime,agSwitchLastBootTime,agSwitchUpTime,agFipsSecurityLevel,agFipsNonApprovedMode,mgmtPortInfoIPv6SLAACTot,agMgmtCurCfgIpAddr,agMgmtCurCfgMask,agMgmtCurCfgGateway,agMgmtCurCfgIpv6Addr,agMgmtCurCfgIpv6PrefixLen,agMgmtCurCfgIpv6Gateway,mgmtPortInfoIPv6SLAAC1Addr,mgmtPortInfoIPv6SLAAC1PrefixLen,mgmtPortInfoIPv6SLAAC2Addr,mgmtPortInfoIPv6SLAAC2PrefixLen,mgmtPortInfoIPv6SLAAC3Addr,mgmtPortInfoIPv6SLAAC3PrefixLen,mgmtPortInfoIPv6SLAAC4Addr,mgmtPortInfoIPv6SLAAC4PrefixLen,hwMACAddress"

	body, err := s.makeRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var systemInfo models.SystemInfo
	if err := json.Unmarshal(body, &systemInfo); err != nil {
		return nil, fmt.Errorf("parseando JSON: %w", err)
	}
	return &systemInfo, nil
}

func (s *AlteonService) GetLicenses(ctx context.Context) (*models.LicenseResponse, error) {
	licenseEndpoint := "/config/AgLicenseInfoTable?count=50&props=LicenseInfoIdx,SoftwareKey,TimeBasedLicenseStatus"
	licenseBody, err := s.makeRequest(ctx, licenseEndpoint)
	if err != nil {
		return nil, fmt.Errorf("obteniendo licencias: %w", err)
	}

	var licenseResponse models.AlteonLicenseInfoResponse
	if err := json.Unmarshal(licenseBody, &licenseResponse); err != nil {
		return nil, fmt.Errorf("parseando licencias: %w", err)
	}

	capacityEndpoint := "/config/AgLicenseCapacityInfoTable?count=50&props=LicenseCapacityInfoIdx,LicenseCapacitySize,LicenseCapacityCurrUsage,LicenseCapacityPeakUsage"
	capacityBody, err := s.makeRequest(ctx, capacityEndpoint)
	if err != nil {
		return nil, fmt.Errorf("obteniendo capacidad: %w", err)
	}

	var capacityResponse models.AlteonLicenseCapacityResponse
	if err := json.Unmarshal(capacityBody, &capacityResponse); err != nil {
		return nil, fmt.Errorf("parseando capacidad: %w", err)
	}

	capacityMap := make(map[int]models.LicenseCapacityInfo)
	for _, capacity := range capacityResponse.AgLicenseCapacityInfoTable {
		capacityMap[capacity.LicenseCapacityInfoIdx] = capacity
	}

	combinedLicenses := []models.CombinedLicense{}

	for _, license := range licenseResponse.AgLicenseInfoTable {
		if license.SoftwareKey == "" {
			continue
		}

		combined := models.CombinedLicense{
			LicenseIdx:      license.LicenseInfoIdx,
			SoftwareKey:     license.SoftwareKey,
			Status:          license.TimeBasedLicenseStatus,
			HasCapacityInfo: false,
		}

		expirationDate, daysUntil := parseExpirationDate(license.TimeBasedLicenseStatus)
		if expirationDate != "" {
			combined.ExpirationDate = expirationDate
			combined.DaysUntilExpiration = daysUntil
		}

		if capacity, exists := capacityMap[license.LicenseInfoIdx]; exists {
			combined.HasCapacityInfo = true
			combined.CapacitySize = capacity.LicenseCapacitySize
			combined.CapacitySizeFormatted = formatCapacitySize(capacity.LicenseCapacitySize)
			combined.CurrentUsage = capacity.LicenseCapacityCurrUsage
			combined.PeakUsage = capacity.LicenseCapacityPeakUsage
		}

		combinedLicenses = append(combinedLicenses, combined)
	}

	return &models.LicenseResponse{Licenses: combinedLicenses}, nil
}

func (s *AlteonService) GetVirtualServers(ctx context.Context) (*models.VirtualServersResponse, error) {
	vserverEndpoint := "/config/SlbStatEnhVServerTable?count=50&props=Index,SessionsPerSec,OctetsPerSec,CurrSessions,TotalSessions,HighestSessions,HCOctets"
	vserverBody, err := s.makeRequest(ctx, vserverEndpoint)
	if err != nil {
		return nil, fmt.Errorf("obteniendo servidores virtuales: %w", err)
	}

	var vserverResponse models.SlbStatEnhVServerTableResponse
	if err := json.Unmarshal(vserverBody, &vserverResponse); err != nil {
		return nil, fmt.Errorf("parseando servidores virtuales: %w", err)
	}

	s.logCtx(ctx).WithField("count", len(vserverResponse.SlbStatEnhVServerTable)).Debug("virtual servers obtenidos")

	virtualServers := make([]models.VirtualServer, len(vserverResponse.SlbStatEnhVServerTable))

	var wg sync.WaitGroup
	for i, vserver := range vserverResponse.SlbStatEnhVServerTable {
		wg.Add(1)
		go func(idx int, vs models.SlbStatEnhVServer) {
			defer wg.Done()
			virtualServers[idx] = s.buildVirtualServer(ctx, vs)
		}(i, vserver)
	}
	wg.Wait()

	return &models.VirtualServersResponse{VirtualServers: virtualServers}, nil
}

func (s *AlteonService) buildVirtualServer(ctx context.Context, vs models.SlbStatEnhVServer) models.VirtualServer {
	out := models.VirtualServer{
		Index:           vs.Index,
		SessionsPerSec:  vs.SessionsPerSec,
		OctetsPerSec:    vs.OctetsPerSec,
		CurrSessions:    vs.CurrSessions,
		TotalSessions:   vs.TotalSessions,
		HighestSessions: vs.HighestSessions,
		HCOctets:        vs.HCOctets,
		Services:        []models.VirtualService{},
	}

	services, err := s.getVirtualServerServices(ctx, vs.Index)
	if err != nil {
		s.log().WithField("vserver", vs.Index).WithError(err).Warn("servicios del vserver fallaron")
		return out
	}

	out.Services = make([]models.VirtualService, len(services))

	var wg sync.WaitGroup
	for i, svc := range services {
		wg.Add(1)
		go func(idx int, svc models.SlbEnhVirtServicesInfo) {
			defer wg.Done()
			out.Services[idx] = s.buildVirtualService(ctx, svc)
		}(i, svc)
	}
	wg.Wait()

	return out
}

func (s *AlteonService) buildVirtualService(ctx context.Context, svc models.SlbEnhVirtServicesInfo) models.VirtualService {
	vsvc := models.VirtualService{
		VirtServIndex:   svc.VirtServIndex,
		SvcIndex:        svc.SvcIndex,
		RealServIndex:   svc.RealServIndex,
		Vport:           svc.Vport,
		Rport:           svc.Rport,
		State:           svc.State,
		StateName:       getStateName(svc.State),
		ResponseTime:    svc.ResponseTime,
		Weight:          svc.Weight,
		CfgRealHealth:   svc.CfgRealHealth,
		RtRealHealth:    svc.RtRealHealth,
		StateFailReason: svc.StateFailReason,
		RealLogexp:      svc.RealLogexp,
	}

	var stats *models.RealServerStats
	var info *models.SlbEnhRealServerInfo
	var statsErr, infoErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		stats, statsErr = s.getServiceStatistics(ctx, svc.VirtServIndex, svc.SvcIndex, svc.RealServIndex)
	}()
	go func() {
		defer wg.Done()
		info, infoErr = s.getRealServerInfo(ctx, svc.RealServIndex)
	}()
	wg.Wait()

	if statsErr != nil {
		s.log().WithFields(logrus.Fields{
			"vserver": svc.VirtServIndex,
			"svc":     svc.SvcIndex,
			"real":    svc.RealServIndex,
		}).WithError(statsErr).Warn("estadísticas del servicio fallaron")
		return vsvc
	}
	if infoErr != nil {
		s.log().WithField("real", svc.RealServIndex).WithError(infoErr).Warn("info del real server falló")
	} else if info != nil {
		stats.MacAddr = info.MacAddr
		stats.IpAddr = info.IpAddr
		stats.InfoState = info.State
		stats.InfoStateName = getRealServerInfoStateName(info.State)
	}

	vsvc.RealServer = stats
	return vsvc
}

func (s *AlteonService) getVirtualServerServices(ctx context.Context, vserverIndex string) ([]models.SlbEnhVirtServicesInfo, error) {
	endpoint := fmt.Sprintf("/config/SlbEnhVirtServicesInfoTable/%s/", vserverIndex)
	body, err := s.makeRequestCached(ctx, endpoint, statsTTL)
	if err != nil {
		return nil, err
	}

	var response models.SlbEnhVirtServicesInfoTableResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parseando servicios: %w", err)
	}

	s.log().WithFields(logrus.Fields{
		"vserver": vserverIndex,
		"count":   len(response.SlbEnhVirtServicesInfoTable),
	}).Debug("servicios obtenidos")

	return response.SlbEnhVirtServicesInfoTable, nil
}

func (s *AlteonService) getServiceStatistics(ctx context.Context, vserverIndex string, svcIndex int, realServIndex string) (*models.RealServerStats, error) {
	endpoint := fmt.Sprintf("/config/SlbEnhStatVirtServiceTable/%s/%d/%s?props=RealStatus,CurrSessions,TotalSessions,HighestSessions,HcReason,Thruput,TotalBw,PktPerSec,ServerRtt,ServerIndex,Index,RealServerIndex",
		vserverIndex, svcIndex, realServIndex)

	body, err := s.makeRequestCached(ctx, endpoint, statsTTL)
	if err != nil {
		return nil, err
	}

	var response models.SlbEnhStatVirtServiceTableResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parseando estadísticas: %w", err)
	}

	if len(response.SlbEnhStatVirtServiceTable) == 0 {
		return nil, fmt.Errorf("no se encontraron estadísticas")
	}

	stat := response.SlbEnhStatVirtServiceTable[0]

	return &models.RealServerStats{
		RealStatus:      stat.RealStatus,
		RealStatusName:  getRealStatusName(stat.RealStatus),
		CurrSessions:    stat.CurrSessions,
		TotalSessions:   stat.TotalSessions,
		HighestSessions: stat.HighestSessions,
		HcReason:        stat.HcReason,
		Thruput:         stat.Thruput,
		TotalBw:         stat.TotalBw,
		PktPerSec:       stat.PktPerSec,
		ServerRtt:       stat.ServerRtt,
		ServerIndex:     stat.ServerIndex,
		Index:           stat.Index,
		RealServerIndex: stat.RealServerIndex,
	}, nil
}

func (s *AlteonService) getRealServerInfo(ctx context.Context, realServIndex string) (*models.SlbEnhRealServerInfo, error) {
	endpoint := fmt.Sprintf("/config/SlbEnhRealServerInfoTable/%s?props=State,MacAddr,Index,IpAddr", realServIndex)

	body, err := s.makeRequestCached(ctx, endpoint, statsTTL)
	if err != nil {
		return nil, err
	}

	var response models.SlbEnhRealServerInfoTableResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parseando información del real server: %w", err)
	}

	if len(response.SlbEnhRealServerInfoTable) == 0 {
		return nil, fmt.Errorf("no se encontró información del real server")
	}

	return &response.SlbEnhRealServerInfoTable[0], nil
}

func (s *AlteonService) GetMonitoring(ctx context.Context) (*models.MonitoringResponse, error) {
	cpuMemEndpoint := "/config?prop=mpCpuStatsUtil1Second,mpCpuStatsUtil4Seconds,mpCpuStatsUtil64Seconds,systemMemStatsTotalMemory,systemMemStatsInitConfigMemory,systemMemStatsSafetyMargin1,systemMemStatsSafetyMargin2"
	cpuMemBody, err := s.makeRequest(ctx, cpuMemEndpoint)
	if err != nil {
		return nil, fmt.Errorf("obteniendo estadísticas de CPU y memoria: %w", err)
	}

	var cpuMemStats models.CPUMemoryStatsResponse
	if err := json.Unmarshal(cpuMemBody, &cpuMemStats); err != nil {
		return nil, fmt.Errorf("parseando estadísticas de CPU y memoria: %w", err)
	}

	memCoreEndpoint := "/config/SpMemUseStatsTable?count=50&props=Index,InitSizeTo1Margin,InitSizeTo2Margin,CurProcSize,CurProcCacheRelativeSize,CurProcDynCertRelativeSize,CurExtraProcessRelativeSize,CurQatSlabsRelativeSize,MemPressStat,MemPressActiveTime,MemUseFrom1stMargin,PeakUsageFrom1stMargin"
	memCoreBody, err := s.makeRequest(ctx, memCoreEndpoint)
	if err != nil {
		return nil, fmt.Errorf("obteniendo estadísticas de memoria por core: %w", err)
	}

	var memCoreStats models.SpMemUseStatsTableResponse
	if err := json.Unmarshal(memCoreBody, &memCoreStats); err != nil {
		return nil, fmt.Errorf("parseando estadísticas de memoria por core: %w", err)
	}

	cpu := models.CPUStats{
		Util1Second:   cpuMemStats.MpCpuStatsUtil1Second,
		Util4Seconds:  cpuMemStats.MpCpuStatsUtil4Seconds,
		Util64Seconds: cpuMemStats.MpCpuStatsUtil64Seconds,
	}

	usedMemory := cpuMemStats.SystemMemStatsInitConfigMemory
	availableMemory := cpuMemStats.SystemMemStatsTotalMemory - usedMemory
	usagePercentage := 0.0
	if cpuMemStats.SystemMemStatsTotalMemory > 0 {
		usagePercentage = (float64(usedMemory) / float64(cpuMemStats.SystemMemStatsTotalMemory)) * 100
	}

	memory := models.MemoryStats{
		TotalMemory:      cpuMemStats.SystemMemStatsTotalMemory,
		InitConfigMemory: cpuMemStats.SystemMemStatsInitConfigMemory,
		SafetyMargin1:    cpuMemStats.SystemMemStatsSafetyMargin1,
		SafetyMargin2:    cpuMemStats.SystemMemStatsSafetyMargin2,
		UsedMemory:       usedMemory,
		AvailableMemory:  availableMemory,
		UsagePercentage:  usagePercentage,
	}

	cores := []models.CoreMemory{}
	for _, core := range memCoreStats.SpMemUseStatsTable {
		cores = append(cores, models.CoreMemory{
			Index:                       core.Index,
			InitSizeTo1Margin:           core.InitSizeTo1Margin,
			InitSizeTo2Margin:           core.InitSizeTo2Margin,
			CurProcSize:                 core.CurProcSize,
			CurProcCacheRelativeSize:    core.CurProcCacheRelativeSize,
			CurProcDynCertRelativeSize:  core.CurProcDynCertRelativeSize,
			CurExtraProcessRelativeSize: core.CurExtraProcessRelativeSize,
			CurQatSlabsRelativeSize:     core.CurQatSlabsRelativeSize,
			MemPressStat:                core.MemPressStat,
			MemPressActiveTime:          core.MemPressActiveTime,
			MemUseFrom1stMargin:         core.MemUseFrom1stMargin,
			PeakUsageFrom1stMargin:      core.PeakUsageFrom1stMargin,
		})
	}

	return &models.MonitoringResponse{
		CPU:    cpu,
		Memory: memory,
		Cores:  cores,
	}, nil
}

func (s *AlteonService) GetServiceMap(ctx context.Context) (*models.ServiceMapResponse, error) {
	endpoint := "/monitor/servicemap"
	maxRetries := 8
	retryDelay := 2 * time.Second

	var lastErr error
	var serviceMap models.ServiceMapResponse

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			s.logCtx(ctx).WithFields(logrus.Fields{
				"attempt": attempt + 1,
				"max":     maxRetries,
			}).Debug("reintento service map")
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt)):
			}
		}

		body, err := s.makeRequest(ctx, endpoint)
		if err != nil {
			lastErr = err
			continue
		}

		if err := json.Unmarshal(body, &serviceMap); err != nil {
			lastErr = fmt.Errorf("parseando service map: %w", err)
			continue
		}

		if serviceMap.Status == "err" {
			s.logCtx(ctx).Debug("service map devolvió status err, reintentando")
			lastErr = fmt.Errorf("service map status: err")
			continue
		}

		cleanServiceMap(&serviceMap)
		s.logCtx(ctx).WithField("attempt", attempt+1).Debug("service map obtenido")
		return &serviceMap, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("service map falló después de %d intentos: %w", maxRetries, lastErr)
	}
	return nil, fmt.Errorf("service map: statdb no listo después de %d intentos", maxRetries)
}

func parseExpirationDate(status string) (string, int) {
	if strings.Contains(status, "Expires on") {
		parts := strings.Split(status, "Expires on ")
		if len(parts) == 2 {
			dateStr := strings.TrimSpace(parts[1])
			expirationDate, err := time.Parse("01/02/06", dateStr)
			if err == nil {
				daysUntil := int(expirationDate.Sub(time.Now()).Hours() / 24)
				return dateStr, daysUntil
			}
		}
	}
	return "", 0
}

func formatCapacitySize(size int) string {
	switch {
	case size == -1:
		return "Unlimited"
	case size == 0:
		return "Not Applicable"
	case size >= 1000:
		return fmt.Sprintf("%d Gbps", size/1000)
	default:
		return fmt.Sprintf("%d Mbps", size)
	}
}

func getStateName(state int) string {
	stateNames := map[int]string{
		1: "Blocked",
		2: "Running",
		3: "Failed",
		4: "Disabled",
		5: "Slowstart",
	}
	if name, exists := stateNames[state]; exists {
		return name
	}
	return "Unknown"
}

func getRealStatusName(status int) string {
	statusNames := map[int]string{
		1: "Running",
		2: "Failed",
		3: "Disabled",
		4: "Blocked",
	}
	if name, exists := statusNames[status]; exists {
		return name
	}
	return "Unknown"
}

func getRealServerInfoStateName(state int) string {
	stateNames := map[int]string{
		1: "Disabled",
		2: "Enabled",
	}
	if name, exists := stateNames[state]; exists {
		return name
	}
	return "Unknown"
}

func cleanServiceMap(sm *models.ServiceMapResponse) {
	if sm.Status == "ok" {
		sm.Status = ""
	}

	for i := range sm.VServers {
		for j := range sm.VServers[i].VServices {
			vservice := &sm.VServers[i].VServices[j]

			if vservice.CStatus == "OK" {
				vservice.CStatus = ""
			}

			if vservice.RGroup != nil {
				if vservice.RGroup.CStatus == "OK" {
					vservice.RGroup.CStatus = ""
				}

				for k := range vservice.RGroup.RServers {
					server := &vservice.RGroup.RServers[k]
					if server.CStatus == "OK" {
						server.CStatus = ""
					}
					if server.HCReason == "Not Available" {
						server.HCReason = ""
					}
				}
			}
		}
	}
}
