package models

type MonitoringResponseWrapper struct {
	AlteonName string       `json:"alteonName"`
	AlteonURL  string       `json:"alteonUrl"`
	AlteonIP   string       `json:"alteonIp"`
	CPU        CPUStats     `json:"cpu"`
	Memory     MemoryStats  `json:"memory"`
	Cores      []CoreMemory `json:"cores"`
}

type MonitoringResponse struct {
	CPU    CPUStats     `json:"cpu"`
	Memory MemoryStats  `json:"memory"`
	Cores  []CoreMemory `json:"cores"`
}

type CPUStats struct {
	Util1Second   int `json:"util1Second"`
	Util4Seconds  int `json:"util4Seconds"`
	Util64Seconds int `json:"util64Seconds"`
}

type MemoryStats struct {
	TotalMemory int `json:"totalMemory"`
}

type CoreMemory struct {
	Index                  int `json:"index"`
	MemUseFrom1stMargin    int `json:"memUseFrom1stMargin"`
	PeakUsageFrom1stMargin int `json:"peakUsageFrom1stMargin"`
	// Uso de CPU por core (SpStatsCpuUtilTable, unido por índice de SP)
	Util1Second   int `json:"util1Second"`
	Util4Seconds  int `json:"util4Seconds"`
	Util64Seconds int `json:"util64Seconds"`
}

// Estructuras para las respuestas de Alteon
type CPUMemoryStatsResponse struct {
	MpCpuStatsUtil1Second          int `json:"mpCpuStatsUtil1Second"`
	MpCpuStatsUtil4Seconds         int `json:"mpCpuStatsUtil4Seconds"`
	MpCpuStatsUtil64Seconds        int `json:"mpCpuStatsUtil64Seconds"`
	SystemMemStatsTotalMemory      int `json:"systemMemStatsTotalMemory"`
	SystemMemStatsInitConfigMemory int `json:"systemMemStatsInitConfigMemory"`
	SystemMemStatsSafetyMargin1    int `json:"systemMemStatsSafetyMargin1"`
	SystemMemStatsSafetyMargin2    int `json:"systemMemStatsSafetyMargin2"`
	MpMemStatsTotal                int `json:"mpMemStatsTotal"`
	MpMemStatsFree                 int `json:"mpMemStatsFree"`
}

type SpMemUseStatsTableResponse struct {
	SpMemUseStatsTable []SpMemUseStats `json:"SpMemUseStatsTable"`
}

type SpMemUseStats struct {
	Index                       int `json:"Index"`
	InitSizeTo1Margin           int `json:"InitSizeTo1Margin"`
	InitSizeTo2Margin           int `json:"InitSizeTo2Margin"`
	CurProcSize                 int `json:"CurProcSize"`
	CurProcCacheRelativeSize    int `json:"CurProcCacheRelativeSize"`
	CurProcDynCertRelativeSize  int `json:"CurProcDynCertRelativeSize"`
	CurExtraProcessRelativeSize int `json:"CurExtraProcessRelativeSize"`
	CurQatSlabsRelativeSize     int `json:"CurQatSlabsRelativeSize"`
	MemPressStat                int `json:"MemPressStat"`
	MemPressActiveTime          int `json:"MemPressActiveTime"`
	MemUseFrom1stMargin         int `json:"MemUseFrom1stMargin"`
	PeakUsageFrom1stMargin      int `json:"PeakUsageFrom1stMargin"`
}

// CPU por core (String Processors). Se une a CoreMemory por SpIndex == Index.
type SpStatsCpuUtilTableResponse struct {
	SpStatsCpuUtilTable []SpStatsCpuUtil `json:"SpStatsCpuUtilTable"`
}

type SpStatsCpuUtil struct {
	SpIndex       int `json:"SpIndex"`
	Util1Second   int `json:"Util1Second"`
	Util4Seconds  int `json:"Util4Seconds"`
	Util64Seconds int `json:"Util64Seconds"`
}
