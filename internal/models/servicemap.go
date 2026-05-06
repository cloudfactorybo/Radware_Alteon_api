package models

// ServiceMapResponseWrapper para respuesta de múltiples Alteons
type ServiceMapResponseWrapper struct {
	AlteonName string       `json:"alteonName"`
	AlteonURL  string       `json:"alteonUrl"`
	AlteonIP   string       `json:"alteonIp"`
	Timestamp  int64        `json:"timestamp"`
	VServers   []VServerMap `json:"vservers"`
	Status     string       `json:"status,omitempty"` // Solo si NO es "ok"
}

// ServiceMapResponse para respuesta del Alteon
type ServiceMapResponse struct {
	Timestamp int64        `json:"timestamp"`
	VServers  []VServerMap `json:"vservers"`
	Status    string       `json:"status,omitempty"`
}

// VServerMap representa un Virtual Server en el mapa
type VServerMap struct {
	ID        string        `json:"id"`
	IP        string        `json:"ip"`
	VServices []VServiceMap `json:"vservices"`
}

// VServiceMap representa un Virtual Service (puerto)
type VServiceMap struct {
	Name        string     `json:"name"`
	Action      string     `json:"action"`
	VPort       int        `json:"vport"`
	Application string     `json:"application"`
	Protocol    string     `json:"protocol"`
	CStatus     string     `json:"cstatus,omitempty"` // Solo si NO es "OK"
	RGroup      *RGroupMap `json:"rgroup,omitempty"`
}

// RGroupMap representa un Server Group
type RGroupMap struct {
	ID       string       `json:"id"`
	RServers []RServerMap `json:"rservers"`
	CStatus  string       `json:"cstatus,omitempty"` // Solo si NO es "OK"
}

// RServerMap representa un Real Server
type RServerMap struct {
	ID       string `json:"id"`
	IP       string `json:"ip"`
	RPorts   []int  `json:"rports,omitempty"`    // Omitir si vacío
	CStatus  string `json:"cstatus,omitempty"`   // Solo si NO es "OK"
	HCReason string `json:"hc_reason,omitempty"` // Solo si NO es "Not Available"
}
