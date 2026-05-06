package models

type VirtualServersResponseWrapper struct {
	AlteonName     string          `json:"alteonName"`
	AlteonURL      string          `json:"alteonUrl"`
	AlteonIP       string          `json:"alteonIp"`
	VirtualServers []VirtualServer `json:"virtualServers"`
}

type VirtualServersResponse struct {
	VirtualServers []VirtualServer `json:"virtualServers"`
}

type VirtualServer struct {
	Index           string           `json:"index"`
	SessionsPerSec  int              `json:"sessionsPerSec"`
	OctetsPerSec    string           `json:"octetsPerSec"`
	CurrSessions    int              `json:"currSessions"`
	TotalSessions   int              `json:"totalSessions"`
	HighestSessions int              `json:"highestSessions"`
	HCOctets        int              `json:"hcOctets"`
	Services        []VirtualService `json:"services"`
}

type VirtualService struct {
	VirtServIndex   string           `json:"virtServIndex"`
	SvcIndex        int              `json:"svcIndex"`
	RealServIndex   string           `json:"realServIndex"`
	Vport           int              `json:"vport"`
	Rport           int              `json:"rport"`
	State           int              `json:"state"`
	StateName       string           `json:"stateName"`
	ResponseTime    int              `json:"responseTime"`
	Weight          int              `json:"weight"`
	CfgRealHealth   string           `json:"cfgRealHealth"`
	RtRealHealth    string           `json:"rtRealHealth"`
	StateFailReason string           `json:"stateFailReason"`
	RealLogexp      string           `json:"realLogexp"`
	RealServer      *RealServerStats `json:"real_server,omitempty"`
}

type RealServerStats struct {
	RealStatus      int    `json:"realStatus"`
	RealStatusName  string `json:"realStatusName"`
	CurrSessions    int    `json:"currSessions"`
	TotalSessions   int    `json:"totalSessions"`
	HighestSessions int    `json:"highestSessions"`
	HcReason        string `json:"hcReason"`
	Thruput         int        `json:"thruput"`
	TotalBw         FlexString `json:"totalBw"`
	PktPerSec       int        `json:"pktPerSec"`
	ServerRtt       FlexString `json:"serverRtt"`
	ServerIndex     string `json:"serverIndex"`
	Index           int    `json:"index"`
	RealServerIndex string `json:"realServerIndex"`
	// Nuevos campos
	MacAddr       string `json:"macAddr,omitempty"`
	IpAddr        string `json:"ipAddr,omitempty"`
	InfoState     int    `json:"infoState,omitempty"`
	InfoStateName string `json:"infoStateName,omitempty"`
}

// Estructuras para las respuestas de Alteon
type SlbStatEnhVServerTableResponse struct {
	SlbStatEnhVServerTable []SlbStatEnhVServer `json:"SlbStatEnhVServerTable"`
}

type SlbStatEnhVServer struct {
	Index           string `json:"Index"`
	SessionsPerSec  int    `json:"SessionsPerSec"`
	OctetsPerSec    string `json:"OctetsPerSec"`
	CurrSessions    int    `json:"CurrSessions"`
	TotalSessions   int    `json:"TotalSessions"`
	HighestSessions int    `json:"HighestSessions"`
	HCOctets        int    `json:"HCOctets"`
}

type SlbEnhVirtServicesInfoTableResponse struct {
	SlbEnhVirtServicesInfoTable []SlbEnhVirtServicesInfo `json:"SlbEnhVirtServicesInfoTable"`
}

type SlbEnhVirtServicesInfo struct {
	VirtServIndex   string `json:"VirtServIndex"`
	SvcIndex        int    `json:"SvcIndex"`
	RealServIndex   string `json:"RealServIndex"`
	Vport           int    `json:"Vport"`
	Rport           int    `json:"Rport"`
	State           int    `json:"State"`
	ResponseTime    int    `json:"ResponseTime"`
	Weight          int    `json:"Weight"`
	CfgRealHealth   string `json:"CfgRealHealth"`
	RtRealHealth    string `json:"RtRealHealth"`
	StateFailReason string `json:"StateFailReason"`
	RealLogexp      string `json:"RealLogexp"`
}

type SlbEnhStatVirtServiceTableResponse struct {
	SlbEnhStatVirtServiceTable []SlbEnhStatVirtService `json:"SlbEnhStatVirtServiceTable"`
}

type SlbEnhStatVirtService struct {
	RealStatus      int    `json:"RealStatus"`
	CurrSessions    int    `json:"CurrSessions"`
	TotalSessions   int    `json:"TotalSessions"`
	HighestSessions int    `json:"HighestSessions"`
	HcReason        string `json:"HcReason"`
	Thruput         int        `json:"Thruput"`
	TotalBw         FlexString `json:"TotalBw"`
	PktPerSec       int        `json:"PktPerSec"`
	ServerRtt       FlexString `json:"ServerRtt"`
	ServerIndex     string `json:"ServerIndex"`
	Index           int    `json:"Index"`
	RealServerIndex string `json:"RealServerIndex"`
}

// Nueva estructura para la información del Real Server
type SlbEnhRealServerInfoTableResponse struct {
	SlbEnhRealServerInfoTable []SlbEnhRealServerInfo `json:"SlbEnhRealServerInfoTable"`
}

type SlbEnhRealServerInfo struct {
	State   int    `json:"State"`
	MacAddr string `json:"MacAddr"`
	Index   string `json:"Index"`
	IpAddr  string `json:"IpAddr,omitempty"`
}
