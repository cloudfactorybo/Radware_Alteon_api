package service

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"alteon-api/internal/models"
)

type AlteonService struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	logger     *logrus.Logger
}

func NewAlteonService(baseURL, username, password string, client *http.Client, logger *logrus.Logger) *AlteonService {
	customClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 0,
	}

	return &AlteonService{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: customClient,
		logger:     logger,
	}
}

func (s *AlteonService) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", s.baseURL, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	s.logger.Debugf("Realizando petición a: %s", url)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Errorf("Error ejecutando request: %v", err)
		return nil, fmt.Errorf("error ejecutando request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode == 401 {
		s.logger.Errorf("Response body: %s", string(body))
		return nil, fmt.Errorf("error de autenticación 401 - verifica usuario y contraseña")
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("Response body: %s", string(body))
		return nil, fmt.Errorf("error en respuesta: status code %d", resp.StatusCode)
	}

	return body, nil
}

func (s *AlteonService) GetSystemInfo() (*models.SystemInfo, error) {
	endpoint := "/config?prop=sysName,agRtcTime,mpMemStatsFree,agRtcDate,mpMemStatsTotal,agSwitchLastApplyTime,agSwitchLastSaveTime,agSwitchLastBootTime,agSwitchUpTime,agFipsSecurityLevel,agFipsNonApprovedMode,mgmtPortInfoIPv6SLAACTot,agMgmtCurCfgIpAddr,agMgmtCurCfgMask,agMgmtCurCfgGateway,agMgmtCurCfgIpv6Addr,agMgmtCurCfgIpv6PrefixLen,agMgmtCurCfgIpv6Gateway,mgmtPortInfoIPv6SLAAC1Addr,mgmtPortInfoIPv6SLAAC1PrefixLen,mgmtPortInfoIPv6SLAAC2Addr,mgmtPortInfoIPv6SLAAC2PrefixLen,mgmtPortInfoIPv6SLAAC3Addr,mgmtPortInfoIPv6SLAAC3PrefixLen,mgmtPortInfoIPv6SLAAC4Addr,mgmtPortInfoIPv6SLAAC4PrefixLen,hwMACAddress"

	body, err := s.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var systemInfo models.SystemInfo
	if err := json.Unmarshal(body, &systemInfo); err != nil {
		return nil, fmt.Errorf("error parseando JSON: %w", err)
	}

	return &systemInfo, nil
}

func (s *AlteonService) GetLicenses() (*models.LicenseResponse, error) {
	// Obtener información de licencias
	licenseEndpoint := "/config/AgLicenseInfoTable?count=50&props=LicenseInfoIdx,SoftwareKey,TimeBasedLicenseStatus"
	licenseBody, err := s.makeRequest(licenseEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo licencias: %w", err)
	}

	var licenseResponse models.AlteonLicenseInfoResponse
	if err := json.Unmarshal(licenseBody, &licenseResponse); err != nil {
		return nil, fmt.Errorf("error parseando licencias: %w", err)
	}

	// Obtener capacidad de licencias
	capacityEndpoint := "/config/AgLicenseCapacityInfoTable?count=50&props=LicenseCapacityInfoIdx,LicenseCapacitySize,LicenseCapacityCurrUsage,LicenseCapacityPeakUsage"
	capacityBody, err := s.makeRequest(capacityEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo capacidad: %w", err)
	}

	var capacityResponse models.AlteonLicenseCapacityResponse
	if err := json.Unmarshal(capacityBody, &capacityResponse); err != nil {
		return nil, fmt.Errorf("error parseando capacidad: %w", err)
	}

	// Crear mapa de capacidades por índice
	capacityMap := make(map[int]models.LicenseCapacityInfo)
	for _, capacity := range capacityResponse.AgLicenseCapacityInfoTable {
		capacityMap[capacity.LicenseCapacityInfoIdx] = capacity
	}

	// Combinar licencias con capacidades
	combinedLicenses := []models.CombinedLicense{}

	for _, license := range licenseResponse.AgLicenseInfoTable {
		// Solo agregar licencias con SoftwareKey válido
		if license.SoftwareKey == "" {
			continue
		}

		combined := models.CombinedLicense{
			LicenseIdx:      license.LicenseInfoIdx,
			SoftwareKey:     license.SoftwareKey,
			Status:          license.TimeBasedLicenseStatus,
			HasCapacityInfo: false,
		}

		// Parsear fecha de expiración
		expirationDate, daysUntil := parseExpirationDate(license.TimeBasedLicenseStatus)
		if expirationDate != "" {
			combined.ExpirationDate = expirationDate
			combined.DaysUntilExpiration = daysUntil
		}

		// Buscar información de capacidad correspondiente
		if capacity, exists := capacityMap[license.LicenseInfoIdx]; exists {
			combined.HasCapacityInfo = true
			combined.CapacitySize = capacity.LicenseCapacitySize
			combined.CapacitySizeFormatted = formatCapacitySize(capacity.LicenseCapacitySize)
			combined.CurrentUsage = capacity.LicenseCapacityCurrUsage
			combined.PeakUsage = capacity.LicenseCapacityPeakUsage
		}

		combinedLicenses = append(combinedLicenses, combined)
	}

	response := &models.LicenseResponse{
		Licenses: combinedLicenses,
	}

	return response, nil
}

func (s *AlteonService) GetVirtualServers() (*models.VirtualServersResponse, error) {
	// Paso 1: Obtener la lista de servidores virtuales
	vserverEndpoint := "/config/SlbStatEnhVServerTable?count=50&props=Index,SessionsPerSec,OctetsPerSec,CurrSessions,TotalSessions,HighestSessions,HCOctets"
	vserverBody, err := s.makeRequest(vserverEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo servidores virtuales: %w", err)
	}

	var vserverResponse models.SlbStatEnhVServerTableResponse
	if err := json.Unmarshal(vserverBody, &vserverResponse); err != nil {
		return nil, fmt.Errorf("error parseando servidores virtuales: %w", err)
	}

	s.logger.Infof("Se encontraron %d servidores virtuales", len(vserverResponse.SlbStatEnhVServerTable))

	// Construir la respuesta combinada
	virtualServers := []models.VirtualServer{}

	for _, vserver := range vserverResponse.SlbStatEnhVServerTable {
		virtualServer := models.VirtualServer{
			Index:           vserver.Index,
			SessionsPerSec:  vserver.SessionsPerSec,
			OctetsPerSec:    vserver.OctetsPerSec,
			CurrSessions:    vserver.CurrSessions,
			TotalSessions:   vserver.TotalSessions,
			HighestSessions: vserver.HighestSessions,
			HCOctets:        vserver.HCOctets,
			Services:        []models.VirtualService{},
		}

		// Paso 2: Obtener los servicios de cada servidor virtual
		services, err := s.getVirtualServerServices(vserver.Index)
		if err != nil {
			s.logger.Errorf("Error obteniendo servicios del servidor virtual %s: %v", vserver.Index, err)
			// Continuar con el siguiente servidor virtual aunque falle uno
			continue
		}

		// Paso 3: Para cada servicio, obtener sus estadísticas
		for _, service := range services {
			virtualService := models.VirtualService{
				VirtServIndex:   service.VirtServIndex,
				SvcIndex:        service.SvcIndex,
				RealServIndex:   service.RealServIndex,
				Vport:           service.Vport,
				Rport:           service.Rport,
				State:           service.State,
				StateName:       getStateName(service.State),
				ResponseTime:    service.ResponseTime,
				Weight:          service.Weight,
				CfgRealHealth:   service.CfgRealHealth,
				RtRealHealth:    service.RtRealHealth,
				StateFailReason: service.StateFailReason,
				RealLogexp:      service.RealLogexp,
			}

			// Obtener estadísticas del servicio
			stats, err := s.getServiceStatistics(service.VirtServIndex, service.SvcIndex, service.RealServIndex)
			if err != nil {
				s.logger.Warnf("Error obteniendo estadísticas del servicio %s/%d/%s: %v",
					service.VirtServIndex, service.SvcIndex, service.RealServIndex, err)
				// Continuar sin estadísticas
			} else {
				// Obtener información adicional del Real Server (MAC e IP)
				realServerInfo, err := s.getRealServerInfo(service.RealServIndex)
				if err != nil {
					s.logger.Warnf("Error obteniendo información del real server %s: %v", service.RealServIndex, err)
				} else {
					// Agregar MAC, IP y State del Real Server a las estadísticas
					stats.MacAddr = realServerInfo.MacAddr
					stats.IpAddr = realServerInfo.IpAddr
					stats.InfoState = realServerInfo.State
					stats.InfoStateName = getRealServerInfoStateName(realServerInfo.State)
				}

				virtualService.RealServer = stats
			}

			virtualServer.Services = append(virtualServer.Services, virtualService)
		}

		virtualServers = append(virtualServers, virtualServer)
	}

	response := &models.VirtualServersResponse{
		VirtualServers: virtualServers,
	}

	return response, nil
}

func (s *AlteonService) getVirtualServerServices(vserverIndex string) ([]models.SlbEnhVirtServicesInfo, error) {
	endpoint := fmt.Sprintf("/config/SlbEnhVirtServicesInfoTable/%s/", vserverIndex)
	body, err := s.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response models.SlbEnhVirtServicesInfoTableResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parseando servicios: %w", err)
	}

	s.logger.Debugf("Servidor virtual %s tiene %d servicios", vserverIndex, len(response.SlbEnhVirtServicesInfoTable))

	return response.SlbEnhVirtServicesInfoTable, nil
}

func (s *AlteonService) getServiceStatistics(vserverIndex string, svcIndex int, realServIndex string) (*models.RealServerStats, error) {
	endpoint := fmt.Sprintf("/config/SlbEnhStatVirtServiceTable/%s/%d/%s?props=RealStatus,CurrSessions,TotalSessions,HighestSessions,HcReason,Thruput,TotalBw,PktPerSec,ServerRtt,ServerIndex,Index,RealServerIndex",
		vserverIndex, svcIndex, realServIndex)

	body, err := s.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response models.SlbEnhStatVirtServiceTableResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parseando estadísticas: %w", err)
	}

	if len(response.SlbEnhStatVirtServiceTable) == 0 {
		return nil, fmt.Errorf("no se encontraron estadísticas")
	}

	stat := response.SlbEnhStatVirtServiceTable[0]

	realServerStats := &models.RealServerStats{
		RealStatus:      stat.RealStatus,
		RealStatusName:  getRealStatusName(stat.RealStatus),
		CurrSessions:    stat.CurrSessions,
		TotalSessions:   stat.TotalSessions,
		HighestSessions: stat.HighestSessions,
		HcReason:        stat.HcReason,
		Thruput:         stat.Thruput,
		TotalBw:         stat.TotalBw, // Ahora es string a string
		PktPerSec:       stat.PktPerSec,
		ServerRtt:       stat.ServerRtt,
		ServerIndex:     stat.ServerIndex,
		Index:           stat.Index,
		RealServerIndex: stat.RealServerIndex,
	}

	return realServerStats, nil
}

// Nueva función para obtener información del Real Server (MAC e IP)
func (s *AlteonService) getRealServerInfo(realServIndex string) (*models.SlbEnhRealServerInfo, error) {
	endpoint := fmt.Sprintf("/config/SlbEnhRealServerInfoTable/%s?props=State,MacAddr,Index,IpAddr", realServIndex)

	body, err := s.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response models.SlbEnhRealServerInfoTableResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parseando información del real server: %w", err)
	}

	if len(response.SlbEnhRealServerInfoTable) == 0 {
		return nil, fmt.Errorf("no se encontró información del real server")
	}

	return &response.SlbEnhRealServerInfoTable[0], nil
}

func (s *AlteonService) GetMonitoring() (*models.MonitoringResponse, error) {
	// Paso 1: Obtener estadísticas de CPU y memoria general
	cpuMemEndpoint := "/config?prop=mpCpuStatsUtil1Second,mpCpuStatsUtil4Seconds,mpCpuStatsUtil64Seconds,systemMemStatsTotalMemory,systemMemStatsInitConfigMemory,systemMemStatsSafetyMargin1,systemMemStatsSafetyMargin2"
	cpuMemBody, err := s.makeRequest(cpuMemEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas de CPU y memoria: %w", err)
	}

	var cpuMemStats models.CPUMemoryStatsResponse
	if err := json.Unmarshal(cpuMemBody, &cpuMemStats); err != nil {
		return nil, fmt.Errorf("error parseando estadísticas de CPU y memoria: %w", err)
	}

	// Paso 2: Obtener estadísticas de memoria por core
	memCoreEndpoint := "/config/SpMemUseStatsTable?count=50&props=Index,InitSizeTo1Margin,InitSizeTo2Margin,CurProcSize,CurProcCacheRelativeSize,CurProcDynCertRelativeSize,CurExtraProcessRelativeSize,CurQatSlabsRelativeSize,MemPressStat,MemPressActiveTime,MemUseFrom1stMargin,PeakUsageFrom1stMargin"
	memCoreBody, err := s.makeRequest(memCoreEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estadísticas de memoria por core: %w", err)
	}

	var memCoreStats models.SpMemUseStatsTableResponse
	if err := json.Unmarshal(memCoreBody, &memCoreStats); err != nil {
		return nil, fmt.Errorf("error parseando estadísticas de memoria por core: %w", err)
	}

	// Construir respuesta combinada
	cpu := models.CPUStats{
		Util1Second:   cpuMemStats.MpCpuStatsUtil1Second,
		Util4Seconds:  cpuMemStats.MpCpuStatsUtil4Seconds,
		Util64Seconds: cpuMemStats.MpCpuStatsUtil64Seconds,
	}

	// Calcular memoria usada y disponible
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

	// Convertir cores
	cores := []models.CoreMemory{}
	for _, core := range memCoreStats.SpMemUseStatsTable {
		coreMemory := models.CoreMemory{
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
		cores = append(cores, coreMemory)
	}

	response := &models.MonitoringResponse{
		CPU:    cpu,
		Memory: memory,
		Cores:  cores,
	}

	return response, nil
}

func (s *AlteonService) GetServiceMap() (*models.ServiceMapResponse, error) {
	endpoint := "/monitor/servicemap"
	maxRetries := 8
	retryDelay := 2 * time.Second

	var lastErr error
	var serviceMap models.ServiceMapResponse

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			s.logger.Infof("Reintento %d/%d para service map", attempt+1, maxRetries)
			time.Sleep(retryDelay * time.Duration(attempt))
		}

		body, err := s.makeRequest(endpoint)
		if err != nil {
			lastErr = err
			continue
		}

		if err := json.Unmarshal(body, &serviceMap); err != nil {
			lastErr = fmt.Errorf("error parseando service map: %w", err)
			continue
		}

		// Verificar si la respuesta es válida (status != "err")
		if serviceMap.Status == "err" {
			s.logger.Infof("Service map devolvió status 'err', reintentando...")
			lastErr = fmt.Errorf("service map status: err")
			continue
		}

		// Limpiar campos redundantes
		cleanServiceMap(&serviceMap)

		s.logger.Infof("Service map obtenido exitosamente en intento %d", attempt+1)
		return &serviceMap, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("error obteniendo service map después de %d intentos: %w", maxRetries, lastErr)
	}

	return nil, fmt.Errorf("error obteniendo service map: statdb not ready después de %d intentos", maxRetries)
}

// parseExpirationDate extrae la fecha de expiración del string "Expires on 10/11/27"
func parseExpirationDate(status string) (string, int) {
	if strings.Contains(status, "Expires on") {
		parts := strings.Split(status, "Expires on ")
		if len(parts) == 2 {
			dateStr := strings.TrimSpace(parts[1])

			// Parsear fecha
			expirationDate, err := time.Parse("01/02/06", dateStr)
			if err == nil {
				// Calcular días hasta expiración
				now := time.Now()
				daysUntil := int(expirationDate.Sub(now).Hours() / 24)
				return dateStr, daysUntil
			}
		}
	}
	return "", 0
}

// formatCapacitySize formatea el tamaño de capacidad
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

// getStateName convierte el código de estado a nombre legible
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

// getRealStatusName convierte el código de estado real a nombre legible
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

// getRealServerInfoStateName convierte el código de estado de información a nombre legible
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

// cleanServiceMap elimina campos con valor "OK", "ok" o "Not Available" para reducir ruido
func cleanServiceMap(sm *models.ServiceMapResponse) {
	// Limpiar status "ok" del response principal
	if sm.Status == "ok" {
		sm.Status = ""
	}

	for i := range sm.VServers {
		for j := range sm.VServers[i].VServices {
			vservice := &sm.VServers[i].VServices[j]

			// Limpiar cstatus "OK"
			if vservice.CStatus == "OK" {
				vservice.CStatus = ""
			}

			if vservice.RGroup != nil {
				// Limpiar cstatus del grupo
				if vservice.RGroup.CStatus == "OK" {
					vservice.RGroup.CStatus = ""
				}

				// Limpiar status de los servidores
				for k := range vservice.RGroup.RServers {
					server := &vservice.RGroup.RServers[k]
					if server.CStatus == "OK" {
						server.CStatus = ""
					}
					// Limpiar health reason si es "Not Available"
					if server.HCReason == "Not Available" {
						server.HCReason = ""
					}
				}
			}
		}
	}
}
