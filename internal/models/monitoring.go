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
	TotalMemory      int     `json:"totalMemory"`
	InitConfigMemory int     `json:"initConfigMemory"`
	SafetyMargin1    int     `json:"safetyMargin1"`
	SafetyMargin2    int     `json:"safetyMargin2"`
	UsedMemory       int     `json:"usedMemory"`
	AvailableMemory  int     `json:"availableMemory"`
	UsagePercentage  float64 `json:"usagePercentage"`
}

type CoreMemory struct {
	Index                       int `json:"index"`
	InitSizeTo1Margin           int `json:"initSizeTo1Margin"`
	InitSizeTo2Margin           int `json:"initSizeTo2Margin"`
	CurProcSize                 int `json:"curProcSize"`
	CurProcCacheRelativeSize    int `json:"curProcCacheRelativeSize"`
	CurProcDynCertRelativeSize  int `json:"curProcDynCertRelativeSize"`
	CurExtraProcessRelativeSize int `json:"curExtraProcessRelativeSize"`
	CurQatSlabsRelativeSize     int `json:"curQatSlabsRelativeSize"`
	MemPressStat                int `json:"memPressStat"`
	MemPressActiveTime          int `json:"memPressActiveTime"`
	MemUseFrom1stMargin         int `json:"memUseFrom1stMargin"`
	PeakUsageFrom1stMargin      int `json:"peakUsageFrom1stMargin"`
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
