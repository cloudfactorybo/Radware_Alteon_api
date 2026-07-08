package models

// ---------- Smart NAT ----------

type SmartNatResponseWrapper struct {
	AlteonName string         `json:"alteonName"`
	AlteonURL  string         `json:"alteonUrl"`
	AlteonIP   string         `json:"alteonIp"`
	Entries    []SmartNatRule `json:"entries"`
}

type SmartNatResponse struct {
	Entries []SmartNatRule `json:"entries"`
}

type SmartNatRule struct {
	ID            string `json:"id"`
	Type          int    `json:"type"`
	CurrSessions  int    `json:"currSessions"`
	TotalSessions int64  `json:"totalSessions"`
	LocalIp       string `json:"localIp"`
	LocalMask     string `json:"localMask"`
	DnatIp        string `json:"dnatIp"`
	DnatMask      string `json:"dnatMask"`
	WanLink       string `json:"wanLink"`
	DnatPersist   int    `json:"dnatPersist"`
}

type SlbStatLinkpfSmartNATTableResponse struct {
	SlbStatLinkpfSmartNATTable []AlteonSmartNatStat `json:"SlbStatLinkpfSmartNATTable"`
}

type AlteonSmartNatStat struct {
	NATIndex    string `json:"NATIndex"`
	NATCurrSess int    `json:"NATCurrSess"`
	NATTotSess  int64  `json:"NATTotSess"`
	NATType     int    `json:"NATType"`
}

type SlbCurCfgSmartNatTableResponse struct {
	SlbCurCfgSmartNatTable []AlteonSmartNat `json:"SlbCurCfgSmartNatTable"`
}

type AlteonSmartNat struct {
	Index         string `json:"Index"`
	Type          int    `json:"Type"`
	LocalIpV4     string `json:"LocalIpV4"`
	LocalIpV4Mask string `json:"LocalIpV4Mask"`
	DnatIpV4      string `json:"DnatIpV4"`
	DnatIpV4Mask  string `json:"DnatIpV4Mask"`
	WanLink       string `json:"WanLink"`
	DnatPersist   int    `json:"DnatPersist"`
}

// ---------- WAN Link Groups ----------

type WanLinkGroupsResponseWrapper struct {
	AlteonName string         `json:"alteonName"`
	AlteonURL  string         `json:"alteonUrl"`
	AlteonIP   string         `json:"alteonIp"`
	Groups     []WanLinkGroup `json:"groups"`
}

type WanLinkGroupsResponse struct {
	Groups []WanLinkGroup `json:"groups"`
}

type WanLinkGroup struct {
	ID              string  `json:"id"`
	CurrSessions    int     `json:"currSessions"`
	TotalSessions   int64   `json:"totalSessions"`
	HighestSessions int     `json:"highestSessions"`
	HCOctets        int64   `json:"hcOctets"`
	TotalMB         float64 `json:"totalMB"`
	Metric          int     `json:"metric"`
	MetricName      string  `json:"metricName"`
	BackupServer    string  `json:"backupServer,omitempty"`
}

type SlbStatEnhGroupTableResponse struct {
	SlbStatEnhGroupTable []AlteonEnhGroupStat `json:"SlbStatEnhGroupTable"`
}

type AlteonEnhGroupStat struct {
	Index           string `json:"Index"`
	CurrSessions    int    `json:"CurrSessions"`
	TotalSessions   int64  `json:"TotalSessions"`
	HighestSessions int    `json:"HighestSessions"`
	HCOctets        int64  `json:"HCOctets"`
}

type SlbCurCfgEnhGroupTableResponse struct {
	SlbCurCfgEnhGroupTable []AlteonEnhGroupCfg `json:"SlbCurCfgEnhGroupTable"`
}

type AlteonEnhGroupCfg struct {
	Index        string `json:"Index"`
	Metric       int    `json:"Metric"`
	BackupServer string `json:"BackupServer"`
}

// ---------- WAN Links ----------

type WanLinksResponseWrapper struct {
	AlteonName string    `json:"alteonName"`
	AlteonURL  string    `json:"alteonUrl"`
	AlteonIP   string    `json:"alteonIp"`
	PerId      []WanLink `json:"perId"`
	PerIp      []WanLink `json:"perIp"`
}

type WanLinksResponse struct {
	PerId []WanLink `json:"perId"`
	PerIp []WanLink `json:"perIp"`
}

// WanLink representa una fila de las vistas "Per WAN Link ID" / "Per WAN Link IP".
// Anchos de banda en Mbps como string ("--" = sin límite configurado para utilización).
type WanLink struct {
	ID           string `json:"id"`
	IpAddr       string `json:"ipAddr,omitempty"`
	State        int    `json:"state,omitempty"`
	StateName    string `json:"stateName,omitempty"`
	CurrSessions int    `json:"currSessions"`
	UpBwCurr     string `json:"upBwCurr"`
	UpBwUsage    string `json:"upBwUsage"`
	DnBwCurr     string `json:"dnBwCurr"`
	DnBwUsage    string `json:"dnBwUsage"`
	TotBwCurr    string `json:"totBwCurr"`
	TotBwUsage   string `json:"totBwUsage"`
	UpBwPeak     string `json:"upBwPeak"`
	DnBwPeak     string `json:"dnBwPeak"`
	TotBwPeak    string `json:"totBwPeak"`
	UpBwTot      string `json:"upBwTot"`
	DnBwTot      string `json:"dnBwTot"`
	UpDnBwTot    string `json:"upDnBwTot"`
}

// Per WAN Link ID — SlbStatLinkpfRServerTable
type SlbStatLinkpfRServerTableResponse struct {
	SlbStatLinkpfRServerTable []AlteonLinkpfRServer `json:"SlbStatLinkpfRServerTable"`
}

type AlteonLinkpfRServer struct {
	Index        string `json:"Index"`
	IpAddr       string `json:"IpAddr"`
	State        int    `json:"State"`
	CurrSess     int    `json:"CurrSess"`
	UpBwCurr     string `json:"UpBwCurr"`
	UpBwUsage    string `json:"UpBwUsage"`
	DwBwCurr     string `json:"DwBwCurr"`
	DwBwUSage    string `json:"DwBwUSage"`
	TotCurrbw    string `json:"TotCurrbw"`
	TotCurrUsage string `json:"TotCurrUsage"`
	UpBwPeak     string `json:"UpBwPeak"`
	DnBwPeak     string `json:"DnBwPeak"`
	TotBwPeak    string `json:"TotBwPeak"`
	UpBwTot      string `json:"UpBwTot"`
	DnBwTot      string `json:"DnBwTot"`
	UpDnBwTot    string `json:"UpDnBwTot"`
}

// Per WAN Link IP — SlbStatLinkpfIpTable
type SlbStatLinkpfIpTableResponse struct {
	SlbStatLinkpfIpTable []AlteonLinkpfIp `json:"SlbStatLinkpfIpTable"`
}

type AlteonLinkpfIp struct {
	Index         string `json:"Index"`
	CurrSessions  int    `json:"CurrSessions"`
	UpBwCurr      string `json:"UpBwCurr"`
	UpBwCurrUsage string `json:"UpBwCurrUsage"`
	DnBwCurr      string `json:"DnBwCurr"`
	DnBwCurrUsage string `json:"DnBwCurrUsage"`
	TotBwCurr     string `json:"TotBwCurr"`
	TotBwCurrUsage string `json:"TotBwCurrUsage"`
	UpBwPeak      string `json:"UpBwPeak"`
	DnBwPeak      string `json:"DnBwPeak"`
	TotBwPeak     string `json:"TotBwPeak"`
	UpBwTot       string `json:"UpBwTot"`
	DnBwTot       string `json:"DnBwTot"`
	UpDnBwTot     string `json:"UpDnBwTot"`
}
