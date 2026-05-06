package models

type LicenseResponseWrapper struct {
	AlteonName string            `json:"alteonName"`
	AlteonURL  string            `json:"alteonUrl"`
	AlteonIP   string            `json:"alteonIp"`
	Licenses   []CombinedLicense `json:"licenses"`
}

type LicenseResponse struct {
	Licenses []CombinedLicense `json:"licenses"`
}

type LicenseInfo struct {
	LicenseInfoIdx         int    `json:"LicenseInfoIdx"`
	SoftwareKey            string `json:"SoftwareKey"`
	TimeBasedLicenseStatus string `json:"TimeBasedLicenseStatus"`
}

type LicenseCapacityInfo struct {
	LicenseCapacityInfoIdx   int    `json:"LicenseCapacityInfoIdx"`
	LicenseCapacitySize      int    `json:"LicenseCapacitySize"`
	LicenseCapacityCurrUsage string `json:"LicenseCapacityCurrUsage"`
	LicenseCapacityPeakUsage string `json:"LicenseCapacityPeakUsage"`
}

type CombinedLicense struct {
	LicenseIdx            int    `json:"licenseIdx"`
	SoftwareKey           string `json:"softwareKey"`
	Status                string `json:"status"`
	ExpirationDate        string `json:"expirationDate,omitempty"`
	DaysUntilExpiration   int    `json:"daysUntilExpiration,omitempty"`
	CapacitySize          int    `json:"capacitySize,omitempty"`
	CapacitySizeFormatted string `json:"capacitySizeFormatted,omitempty"`
	CurrentUsage          string `json:"currentUsage,omitempty"`
	PeakUsage             string `json:"peakUsage,omitempty"`
	HasCapacityInfo       bool   `json:"hasCapacityInfo"`
}

type AlteonLicenseInfoResponse struct {
	AgLicenseInfoTable []LicenseInfo `json:"AgLicenseInfoTable"`
}

type AlteonLicenseCapacityResponse struct {
	AgLicenseCapacityInfoTable []LicenseCapacityInfo `json:"AgLicenseCapacityInfoTable"`
}
