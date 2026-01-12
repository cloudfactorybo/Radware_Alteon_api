package service

import (
	"alteon-api/internal/models"
)

type AlteonServiceInterface interface {
	GetSystemInfo() (*models.SystemInfo, error)
	GetLicenses() (*models.LicenseResponse, error)
	GetVirtualServers() (*models.VirtualServersResponse, error)
	GetMonitoring() (*models.MonitoringResponse, error)
	GetServiceMap() (*models.ServiceMapResponse, error)
}
