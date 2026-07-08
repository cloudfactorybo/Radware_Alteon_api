package models

type GatewaysResponseWrapper struct {
	AlteonName string      `json:"alteonName"`
	AlteonURL  string      `json:"alteonUrl"`
	AlteonIP   string      `json:"alteonIp"`
	Metric     int         `json:"metric"`
	MetricName string      `json:"metricName"`
	Gateways   []Gateway   `json:"gateways"`
	Interfaces []Interface `json:"interfaces"`
}

type GatewaysResponse struct {
	Metric     int         `json:"metric"`
	MetricName string      `json:"metricName"`
	Gateways   []Gateway   `json:"gateways"`
	Interfaces []Interface `json:"interfaces"`
}

type Gateway struct {
	Index     int    `json:"index"`
	Addr      string `json:"addr"`
	Ipv6Addr  string `json:"ipv6Addr,omitempty"`
	IpVer     int    `json:"ipVer"`
	Interval  int    `json:"interval"`
	Retry     int    `json:"retry"`
	State     int    `json:"state"`
	StateName string `json:"stateName"`
	Arp       int    `json:"arp"`
	ArpName   string `json:"arpName"`
	Vlan      int    `json:"vlan"`
	Priority  int    `json:"priority"`
}

type Interface struct {
	Index       int    `json:"index"`
	Addr        string `json:"addr"`
	Mask        string `json:"mask"`
	Vlan        int    `json:"vlan"`
	State       int    `json:"state"`
	StateName   string `json:"stateName"`
	Peer        string `json:"peer,omitempty"`
	Description string `json:"description,omitempty"`
	IpVer       int    `json:"ipVer"`
}

// Respuestas crudas del Alteon (IpCurCfgGwTable + ipCurCfgGwMetric).
type IpCurCfgGwTableResponse struct {
	IpCurCfgGwTable []AlteonGateway `json:"IpCurCfgGwTable"`
}

type AlteonGateway struct {
	Index    int    `json:"Index"`
	Addr     string `json:"Addr"`
	Interval int    `json:"Interval"`
	Retry    int    `json:"Retry"`
	State    int    `json:"State"`
	Arp      int    `json:"Arp"`
	Vlan     int    `json:"Vlan"`
	Priority int    `json:"Priority"`
	IpVer    int    `json:"IpVer"`
	Ipv6Addr string `json:"Ipv6Addr"`
}

type IpCurCfgGwMetricResponse struct {
	IpCurCfgGwMetric int `json:"ipCurCfgGwMetric"`
}

type IpCurCfgIntfTableResponse struct {
	IpCurCfgIntfTable []AlteonInterface `json:"IpCurCfgIntfTable"`
}

type AlteonInterface struct {
	Index       int    `json:"Index"`
	Addr        string `json:"Addr"`
	Mask        string `json:"Mask"`
	Vlan        int    `json:"Vlan"`
	State       int    `json:"State"`
	Peer        string `json:"Peer"`
	Description string `json:"Description"`
	IpVer       int    `json:"IpVer"`
}
