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
	servicesTTL           = 1 * time.Hour    // cache para la lista de servicios de cada vserver
	vserverBatchSize      = 10               // vservers procesados por lote (throttle al alteon)
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

// GetVirtualServers obtiene los virtual servers. Si indexes tiene elementos,
// solo construye esos (por su Index); vacío = todos. Filtrar es clave para la
// CPU del Alteon: cada vserver dispara N+1 requests anidados (servicios +
// stats + info de real server), así que pedir solo los necesarios reduce
// drásticamente la carga sobre el equipo.
func (s *AlteonService) GetVirtualServers(ctx context.Context, indexes []string) (*models.VirtualServersResponse, error) {
	vserverEndpoint := "/config/SlbStatEnhVServerTable?count=2048&props=Index,SessionsPerSec,OctetsPerSec,CurrSessions,TotalSessions,HighestSessions,HCOctets"
	vserverBody, err := s.makeRequest(ctx, vserverEndpoint)
	if err != nil {
		return nil, fmt.Errorf("obteniendo servidores virtuales: %w", err)
	}

	var vserverResponse models.SlbStatEnhVServerTableResponse
	if err := json.Unmarshal(vserverBody, &vserverResponse); err != nil {
		return nil, fmt.Errorf("parseando servidores virtuales: %w", err)
	}

	var filter map[string]bool
	if len(indexes) > 0 {
		filter = make(map[string]bool, len(indexes))
		for _, idx := range indexes {
			filter[idx] = true
		}
	}

	selected := make([]models.SlbStatEnhVServer, 0, len(vserverResponse.SlbStatEnhVServerTable))
	for _, vs := range vserverResponse.SlbStatEnhVServerTable {
		if filter == nil || filter[vs.Index] {
			selected = append(selected, vs)
		}
	}

	s.logCtx(ctx).WithFields(logrus.Fields{
		"total":    len(vserverResponse.SlbStatEnhVServerTable),
		"selected": len(selected),
		"filtered": filter != nil,
	}).Debug("virtual servers obtenidos")

	virtualServers := make([]models.VirtualServer, len(selected))

	// Paginación interna: se procesan en lotes de vserverBatchSize (10). Se espera
	// a que termine cada lote antes de empezar el siguiente, para no disparar todos
	// los requests anidados a la vez y así limitar la carga sobre la CPU del alteon.
	for start := 0; start < len(selected); start += vserverBatchSize {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		end := start + vserverBatchSize
		if end > len(selected) {
			end = len(selected)
		}

		var wg sync.WaitGroup
		for i := start; i < end; i++ {
			wg.Add(1)
			go func(idx int, vs models.SlbStatEnhVServer) {
				defer wg.Done()
				virtualServers[idx] = s.buildVirtualServer(ctx, vs)
			}(i, selected[i])
		}
		wg.Wait()

		s.logCtx(ctx).WithFields(logrus.Fields{
			"lote_desde": start,
			"lote_hasta": end,
			"total":      len(selected),
		}).Debug("lote de virtual servers procesado")
	}

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
	body, err := s.makeRequestCached(ctx, endpoint, servicesTTL)
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
	cpuMemEndpoint := "/config?prop=mpCpuStatsUtil1Second,mpCpuStatsUtil4Seconds,mpCpuStatsUtil64Seconds,systemMemStatsTotalMemory,systemMemStatsInitConfigMemory,systemMemStatsSafetyMargin1,systemMemStatsSafetyMargin2,mpMemStatsTotal,mpMemStatsFree"
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

	// CPU por core (SpStatsCpuUtilTable). Se une a la memoria por índice de SP.
	// Es best-effort: si falla, los cores quedan sin CPU pero el resto responde.
	cpuByCore := map[int]models.SpStatsCpuUtil{}
	cpuCoreEndpoint := "/config/SpStatsCpuUtilTable?count=50&props=SpIndex,Util1Second,Util4Seconds,Util64Seconds"
	cpuCoreBody, err := s.makeRequest(ctx, cpuCoreEndpoint)
	if err != nil {
		s.logCtx(ctx).WithError(err).Warn("CPU por core falló")
	} else {
		var cpuCoreStats models.SpStatsCpuUtilTableResponse
		if err := json.Unmarshal(cpuCoreBody, &cpuCoreStats); err != nil {
			s.logCtx(ctx).WithError(err).Warn("parseando CPU por core")
		} else {
			for _, c := range cpuCoreStats.SpStatsCpuUtilTable {
				cpuByCore[c.SpIndex] = c
			}
		}
	}

	cpu := models.CPUStats{
		Util1Second:   cpuMemStats.MpCpuStatsUtil1Second,
		Util4Seconds:  cpuMemStats.MpCpuStatsUtil4Seconds,
		Util64Seconds: cpuMemStats.MpCpuStatsUtil64Seconds,
	}

	// Uso real de RAM: mpMemStatsTotal/Free vienen en KB; el resto del shape
	// (totalMemory, usedMemory, availableMemory) se expone en MB.
	totalMB := cpuMemStats.MpMemStatsTotal / 1024
	freeMB := cpuMemStats.MpMemStatsFree / 1024
	usedMemory := totalMB - freeMB
	availableMemory := freeMB
	usagePercentage := 0.0
	if totalMB > 0 {
		usagePercentage = (float64(usedMemory) / float64(totalMB)) * 100
	}

	memory := models.MemoryStats{
		TotalMemory:      totalMB,
		InitConfigMemory: cpuMemStats.SystemMemStatsInitConfigMemory,
		SafetyMargin1:    cpuMemStats.SystemMemStatsSafetyMargin1,
		SafetyMargin2:    cpuMemStats.SystemMemStatsSafetyMargin2,
		UsedMemory:       usedMemory,
		AvailableMemory:  availableMemory,
		UsagePercentage:  usagePercentage,
	}

	cores := []models.CoreMemory{}
	for _, core := range memCoreStats.SpMemUseStatsTable {
		cm := models.CoreMemory{
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
		}
		if cu, ok := cpuByCore[core.Index]; ok {
			cm.Util1Second = cu.Util1Second
			cm.Util4Seconds = cu.Util4Seconds
			cm.Util64Seconds = cu.Util64Seconds
		}
		cores = append(cores, cm)
	}

	return &models.MonitoringResponse{
		CPU:    cpu,
		Memory: memory,
		Cores:  cores,
	}, nil
}

func (s *AlteonService) GetGateways(ctx context.Context) (*models.GatewaysResponse, error) {
	gwBody, err := s.makeRequest(ctx, "/config/IpCurCfgGwTable")
	if err != nil {
		return nil, fmt.Errorf("obteniendo gateways: %w", err)
	}

	var gwResp models.IpCurCfgGwTableResponse
	if err := json.Unmarshal(gwBody, &gwResp); err != nil {
		return nil, fmt.Errorf("parseando gateways: %w", err)
	}

	metric := 0
	metricBody, err := s.makeRequest(ctx, "/config?prop=ipCurCfgGwMetric")
	if err != nil {
		s.logCtx(ctx).WithError(err).Warn("métrica de gateways falló")
	} else {
		var metricResp models.IpCurCfgGwMetricResponse
		if err := json.Unmarshal(metricBody, &metricResp); err != nil {
			s.logCtx(ctx).WithError(err).Warn("parseando métrica de gateways")
		} else {
			metric = metricResp.IpCurCfgGwMetric
		}
	}

	gateways := make([]models.Gateway, 0, len(gwResp.IpCurCfgGwTable))
	for _, gw := range gwResp.IpCurCfgGwTable {
		gateways = append(gateways, models.Gateway{
			Index:     gw.Index,
			Addr:      strings.TrimSpace(gw.Addr),
			Ipv6Addr:  strings.TrimSpace(gw.Ipv6Addr),
			IpVer:     gw.IpVer,
			Interval:  gw.Interval,
			Retry:     gw.Retry,
			State:     gw.State,
			StateName: getGatewayStateName(gw.State),
			Arp:       gw.Arp,
			ArpName:   getGatewayArpName(gw.Arp),
			Vlan:      gw.Vlan,
			Priority:  gw.Priority,
		})
	}

	interfaces, err := s.getInterfaces(ctx)
	if err != nil {
		s.logCtx(ctx).WithError(err).Warn("interfaces fallaron")
		interfaces = []models.Interface{}
	}

	return &models.GatewaysResponse{
		Metric:     metric,
		MetricName: getGatewayMetricName(metric),
		Gateways:   gateways,
		Interfaces: interfaces,
	}, nil
}

func (s *AlteonService) getInterfaces(ctx context.Context) ([]models.Interface, error) {
	body, err := s.makeRequest(ctx, "/config/IpCurCfgIntfTable?count=256")
	if err != nil {
		return nil, err
	}

	var resp models.IpCurCfgIntfTableResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parseando interfaces: %w", err)
	}

	interfaces := make([]models.Interface, 0, len(resp.IpCurCfgIntfTable))
	for _, intf := range resp.IpCurCfgIntfTable {
		peer := strings.TrimSpace(intf.Peer)
		if peer == "0.0.0.0" {
			peer = ""
		}
		interfaces = append(interfaces, models.Interface{
			Index:       intf.Index,
			Addr:        strings.TrimSpace(intf.Addr),
			Mask:        strings.TrimSpace(intf.Mask),
			Vlan:        intf.Vlan,
			State:       intf.State,
			StateName:   getInterfaceStateName(intf.State),
			Peer:        peer,
			Description: strings.TrimSpace(intf.Description),
			IpVer:       intf.IpVer,
		})
	}
	return interfaces, nil
}

func (s *AlteonService) GetSmartNat(ctx context.Context) (*models.SmartNatResponse, error) {
	body, err := s.makeRequest(ctx, "/config/SlbCurCfgSmartNatTable?count=512")
	if err != nil {
		return nil, fmt.Errorf("obteniendo smart nat: %w", err)
	}

	var resp models.SlbCurCfgSmartNatTableResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parseando smart nat: %w", err)
	}

	statByID := map[string]models.AlteonSmartNatStat{}
	statBody, err := s.makeRequest(ctx, "/config/SlbStatLinkpfSmartNATTable?count=512")
	if err != nil {
		s.logCtx(ctx).WithError(err).Warn("sesiones de smart nat fallaron")
	} else {
		var statResp models.SlbStatLinkpfSmartNATTableResponse
		if err := json.Unmarshal(statBody, &statResp); err != nil {
			s.logCtx(ctx).WithError(err).Warn("parseando sesiones de smart nat")
		} else {
			for _, st := range statResp.SlbStatLinkpfSmartNATTable {
				statByID[st.NATIndex] = st
			}
		}
	}

	entries := make([]models.SmartNatRule, 0, len(resp.SlbCurCfgSmartNatTable))
	seen := map[string]bool{}
	for _, e := range resp.SlbCurCfgSmartNatTable {
		id := strings.TrimSpace(e.Index)
		seen[id] = true
		rule := models.SmartNatRule{
			ID:          id,
			Type:        e.Type,
			LocalIp:     strings.TrimSpace(e.LocalIpV4),
			LocalMask:   strings.TrimSpace(e.LocalIpV4Mask),
			DnatIp:      strings.TrimSpace(e.DnatIpV4),
			DnatMask:    strings.TrimSpace(e.DnatIpV4Mask),
			WanLink:     strings.TrimSpace(e.WanLink),
			DnatPersist: e.DnatPersist,
		}
		if st, ok := statByID[id]; ok {
			rule.CurrSessions = st.NATCurrSess
			rule.TotalSessions = st.NATTotSess
		}
		entries = append(entries, rule)
	}
	// Filas de estadística sin config (p.ej. NAT dinámico).
	for _, st := range statByID {
		if seen[st.NATIndex] {
			continue
		}
		entries = append(entries, models.SmartNatRule{
			ID:            st.NATIndex,
			Type:          st.NATType,
			CurrSessions:  st.NATCurrSess,
			TotalSessions: st.NATTotSess,
		})
	}
	return &models.SmartNatResponse{Entries: entries}, nil
}

func (s *AlteonService) GetWanLinkGroups(ctx context.Context) (*models.WanLinkGroupsResponse, error) {
	statBody, err := s.makeRequest(ctx, "/config/SlbStatEnhGroupTable?count=256")
	if err != nil {
		return nil, fmt.Errorf("obteniendo wan link groups: %w", err)
	}

	var statResp models.SlbStatEnhGroupTableResponse
	if err := json.Unmarshal(statBody, &statResp); err != nil {
		return nil, fmt.Errorf("parseando wan link groups: %w", err)
	}

	cfgByIndex := map[string]models.AlteonEnhGroupCfg{}
	cfgBody, err := s.makeRequest(ctx, "/config/SlbCurCfgEnhGroupTable?count=256")
	if err != nil {
		s.logCtx(ctx).WithError(err).Warn("config de wan link groups falló")
	} else {
		var cfgResp models.SlbCurCfgEnhGroupTableResponse
		if err := json.Unmarshal(cfgBody, &cfgResp); err != nil {
			s.logCtx(ctx).WithError(err).Warn("parseando config de wan link groups")
		} else {
			for _, c := range cfgResp.SlbCurCfgEnhGroupTable {
				cfgByIndex[c.Index] = c
			}
		}
	}

	groups := make([]models.WanLinkGroup, 0, len(statResp.SlbStatEnhGroupTable))
	for _, g := range statResp.SlbStatEnhGroupTable {
		group := models.WanLinkGroup{
			ID:              strings.TrimSpace(g.Index),
			CurrSessions:    g.CurrSessions,
			TotalSessions:   g.TotalSessions,
			HighestSessions: g.HighestSessions,
			HCOctets:        g.HCOctets,
			TotalMB:         float64(g.HCOctets) / (1024 * 1024),
		}
		if cfg, ok := cfgByIndex[g.Index]; ok {
			group.Metric = cfg.Metric
			group.MetricName = getGroupMetricName(cfg.Metric)
			group.BackupServer = strings.TrimSpace(cfg.BackupServer)
		}
		groups = append(groups, group)
	}
	return &models.WanLinkGroupsResponse{Groups: groups}, nil
}

func (s *AlteonService) GetWanLinks(ctx context.Context) (*models.WanLinksResponse, error) {
	idBody, err := s.makeRequest(ctx, "/config/SlbStatLinkpfRServerTable?count=256")
	if err != nil {
		return nil, fmt.Errorf("obteniendo wan links (per id): %w", err)
	}
	var idResp models.SlbStatLinkpfRServerTableResponse
	if err := json.Unmarshal(idBody, &idResp); err != nil {
		return nil, fmt.Errorf("parseando wan links (per id): %w", err)
	}

	perId := make([]models.WanLink, 0, len(idResp.SlbStatLinkpfRServerTable))
	for _, l := range idResp.SlbStatLinkpfRServerTable {
		perId = append(perId, models.WanLink{
			ID:           strings.TrimSpace(l.Index),
			IpAddr:       strings.TrimSpace(l.IpAddr),
			State:        l.State,
			StateName:    getRealStatusName(l.State),
			CurrSessions: l.CurrSess,
			UpBwCurr:     strings.TrimSpace(l.UpBwCurr),
			UpBwUsage:    strings.TrimSpace(l.UpBwUsage),
			DnBwCurr:     strings.TrimSpace(l.DwBwCurr),
			DnBwUsage:    strings.TrimSpace(l.DwBwUSage),
			TotBwCurr:    strings.TrimSpace(l.TotCurrbw),
			TotBwUsage:   strings.TrimSpace(l.TotCurrUsage),
			UpBwPeak:     strings.TrimSpace(l.UpBwPeak),
			DnBwPeak:     strings.TrimSpace(l.DnBwPeak),
			TotBwPeak:    strings.TrimSpace(l.TotBwPeak),
			UpBwTot:      strings.TrimSpace(l.UpBwTot),
			DnBwTot:      strings.TrimSpace(l.DnBwTot),
			UpDnBwTot:    strings.TrimSpace(l.UpDnBwTot),
		})
	}

	ipBody, err := s.makeRequest(ctx, "/config/SlbStatLinkpfIpTable?count=256")
	if err != nil {
		return nil, fmt.Errorf("obteniendo wan links (per ip): %w", err)
	}
	var ipResp models.SlbStatLinkpfIpTableResponse
	if err := json.Unmarshal(ipBody, &ipResp); err != nil {
		return nil, fmt.Errorf("parseando wan links (per ip): %w", err)
	}

	perIp := make([]models.WanLink, 0, len(ipResp.SlbStatLinkpfIpTable))
	for _, l := range ipResp.SlbStatLinkpfIpTable {
		perIp = append(perIp, models.WanLink{
			ID:           strings.TrimSpace(l.Index),
			CurrSessions: l.CurrSessions,
			UpBwCurr:     strings.TrimSpace(l.UpBwCurr),
			UpBwUsage:    strings.TrimSpace(l.UpBwCurrUsage),
			DnBwCurr:     strings.TrimSpace(l.DnBwCurr),
			DnBwUsage:    strings.TrimSpace(l.DnBwCurrUsage),
			TotBwCurr:    strings.TrimSpace(l.TotBwCurr),
			TotBwUsage:   strings.TrimSpace(l.TotBwCurrUsage),
			UpBwPeak:     strings.TrimSpace(l.UpBwPeak),
			DnBwPeak:     strings.TrimSpace(l.DnBwPeak),
			TotBwPeak:    strings.TrimSpace(l.TotBwPeak),
			UpBwTot:      strings.TrimSpace(l.UpBwTot),
			DnBwTot:      strings.TrimSpace(l.DnBwTot),
			UpDnBwTot:    strings.TrimSpace(l.UpDnBwTot),
		})
	}

	return &models.WanLinksResponse{PerId: perId, PerIp: perIp}, nil
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

// Enums best-effort para gateways. El REST del vADC no documenta el enum, así que
// se expone también el entero crudo (State/Arp/Metric) por si el nombre no aplica.
func getGatewayStateName(state int) string {
	names := map[int]string{
		2: "Enabled",
		3: "Disabled",
	}
	if name, ok := names[state]; ok {
		return name
	}
	return "Unknown"
}

func getGatewayArpName(arp int) string {
	names := map[int]string{
		2: "Enabled",
		3: "Disabled",
	}
	if name, ok := names[arp]; ok {
		return name
	}
	return "Unknown"
}

func getGatewayMetricName(metric int) string {
	names := map[int]string{
		1: "roundRobin",
		2: "minMisses",
		3: "strict",
	}
	if name, ok := names[metric]; ok {
		return name
	}
	return "Unknown"
}

// Best-effort: enum del metric de grupo SLB. Se expone también el entero crudo.
func getGroupMetricName(metric int) string {
	names := map[int]string{
		1: "roundRobin",
		2: "leastConnections",
		3: "minMisses",
		4: "hash",
		5: "responseTime",
		6: "bandwidth",
		7: "phash",
	}
	if name, ok := names[metric]; ok {
		return name
	}
	return "Unknown"
}

func getInterfaceStateName(state int) string {
	names := map[int]string{
		2: "Enabled",
		3: "Disabled",
	}
	if name, ok := names[state]; ok {
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
