package models

type SystemInfoResponse struct {
	AlteonName                      string      `json:"alteonName"`
	AlteonURL                       string      `json:"alteonUrl"`
	AlteonIP                        string      `json:"alteonIp"`
	SysName                         string      `json:"sysName"`
	AgRtcTime                       string      `json:"agRtcTime"`
	AgRtcDate                       string      `json:"agRtcDate"`
	MpMemStatsFree                  interface{} `json:"mpMemStatsFree"`
	MpMemStatsTotal                 interface{} `json:"mpMemStatsTotal"`
	AgSwitchLastApplyTime           string      `json:"agSwitchLastApplyTime"`
	AgSwitchLastSaveTime            string      `json:"agSwitchLastSaveTime"`
	AgSwitchLastBootTime            string      `json:"agSwitchLastBootTime"`
	AgSwitchUpTime                  string      `json:"agSwitchUpTime"`
	AgFipsSecurityLevel             interface{} `json:"agFipsSecurityLevel"`
	AgFipsNonApprovedMode           interface{} `json:"agFipsNonApprovedMode"`
	MgmtPortInfoIPv6SLAACTot        interface{} `json:"mgmtPortInfoIPv6SLAACTot"`
	AgMgmtCurCfgIpAddr              string      `json:"agMgmtCurCfgIpAddr"`
	AgMgmtCurCfgMask                string      `json:"agMgmtCurCfgMask"`
	AgMgmtCurCfgGateway             string      `json:"agMgmtCurCfgGateway"`
	AgMgmtCurCfgIpv6Addr            string      `json:"agMgmtCurCfgIpv6Addr"`
	AgMgmtCurCfgIpv6PrefixLen       interface{} `json:"agMgmtCurCfgIpv6PrefixLen"`
	AgMgmtCurCfgIpv6Gateway         string      `json:"agMgmtCurCfgIpv6Gateway"`
	MgmtPortInfoIPv6SLAAC1Addr      string      `json:"mgmtPortInfoIPv6SLAAC1Addr"`
	MgmtPortInfoIPv6SLAAC1PrefixLen interface{} `json:"mgmtPortInfoIPv6SLAAC1PrefixLen"`
	MgmtPortInfoIPv6SLAAC2Addr      string      `json:"mgmtPortInfoIPv6SLAAC2Addr"`
	MgmtPortInfoIPv6SLAAC2PrefixLen interface{} `json:"mgmtPortInfoIPv6SLAAC2PrefixLen"`
	MgmtPortInfoIPv6SLAAC3Addr      string      `json:"mgmtPortInfoIPv6SLAAC3Addr"`
	MgmtPortInfoIPv6SLAAC3PrefixLen interface{} `json:"mgmtPortInfoIPv6SLAAC3PrefixLen"`
	MgmtPortInfoIPv6SLAAC4Addr      string      `json:"mgmtPortInfoIPv6SLAAC4Addr"`
	MgmtPortInfoIPv6SLAAC4PrefixLen interface{} `json:"mgmtPortInfoIPv6SLAAC4PrefixLen"`
	HwMACAddress                    string      `json:"hwMACAddress"`
}

type SystemInfo struct {
	SysName                         string      `json:"sysName"`
	AgRtcTime                       string      `json:"agRtcTime"`
	AgRtcDate                       string      `json:"agRtcDate"`
	MpMemStatsFree                  interface{} `json:"mpMemStatsFree"`
	MpMemStatsTotal                 interface{} `json:"mpMemStatsTotal"`
	AgSwitchLastApplyTime           string      `json:"agSwitchLastApplyTime"`
	AgSwitchLastSaveTime            string      `json:"agSwitchLastSaveTime"`
	AgSwitchLastBootTime            string      `json:"agSwitchLastBootTime"`
	AgSwitchUpTime                  string      `json:"agSwitchUpTime"`
	AgFipsSecurityLevel             interface{} `json:"agFipsSecurityLevel"`
	AgFipsNonApprovedMode           interface{} `json:"agFipsNonApprovedMode"`
	MgmtPortInfoIPv6SLAACTot        interface{} `json:"mgmtPortInfoIPv6SLAACTot"`
	AgMgmtCurCfgIpAddr              string      `json:"agMgmtCurCfgIpAddr"`
	AgMgmtCurCfgMask                string      `json:"agMgmtCurCfgMask"`
	AgMgmtCurCfgGateway             string      `json:"agMgmtCurCfgGateway"`
	AgMgmtCurCfgIpv6Addr            string      `json:"agMgmtCurCfgIpv6Addr"`
	AgMgmtCurCfgIpv6PrefixLen       interface{} `json:"agMgmtCurCfgIpv6PrefixLen"`
	AgMgmtCurCfgIpv6Gateway         string      `json:"agMgmtCurCfgIpv6Gateway"`
	MgmtPortInfoIPv6SLAAC1Addr      string      `json:"mgmtPortInfoIPv6SLAAC1Addr"`
	MgmtPortInfoIPv6SLAAC1PrefixLen interface{} `json:"mgmtPortInfoIPv6SLAAC1PrefixLen"`
	MgmtPortInfoIPv6SLAAC2Addr      string      `json:"mgmtPortInfoIPv6SLAAC2Addr"`
	MgmtPortInfoIPv6SLAAC2PrefixLen interface{} `json:"mgmtPortInfoIPv6SLAAC2PrefixLen"`
	MgmtPortInfoIPv6SLAAC3Addr      string      `json:"mgmtPortInfoIPv6SLAAC3Addr"`
	MgmtPortInfoIPv6SLAAC3PrefixLen interface{} `json:"mgmtPortInfoIPv6SLAAC3PrefixLen"`
	MgmtPortInfoIPv6SLAAC4Addr      string      `json:"mgmtPortInfoIPv6SLAAC4Addr"`
	MgmtPortInfoIPv6SLAAC4PrefixLen interface{} `json:"mgmtPortInfoIPv6SLAAC4PrefixLen"`
	HwMACAddress                    string      `json:"hwMACAddress"`
}
