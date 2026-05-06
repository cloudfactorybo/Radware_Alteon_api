package service

import (
	"testing"

	"alteon-api/internal/models"
)

func TestFormatCapacitySize(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{-1, "Unlimited"},
		{0, "Not Applicable"},
		{500, "500 Mbps"},
		{999, "999 Mbps"},
		{1000, "1 Gbps"},
		{5000, "5 Gbps"},
	}
	for _, c := range cases {
		if got := formatCapacitySize(c.in); got != c.want {
			t.Errorf("formatCapacitySize(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseExpirationDate_Valid(t *testing.T) {
	date, days := parseExpirationDate("Expires on 10/11/30")
	if date != "10/11/30" {
		t.Errorf("date = %q, want 10/11/30", date)
	}
	if days <= 0 {
		t.Errorf("days = %d, expected positive for future date", days)
	}
}

func TestParseExpirationDate_Invalid(t *testing.T) {
	cases := []string{
		"",
		"Permanent",
		"Active",
		"Expires on not-a-date",
	}
	for _, in := range cases {
		date, days := parseExpirationDate(in)
		if date != "" || days != 0 {
			t.Errorf("parseExpirationDate(%q) = (%q, %d), want (\"\", 0)", in, date, days)
		}
	}
}

func TestGetStateName(t *testing.T) {
	cases := map[int]string{
		1:   "Blocked",
		2:   "Running",
		3:   "Failed",
		4:   "Disabled",
		5:   "Slowstart",
		999: "Unknown",
	}
	for in, want := range cases {
		if got := getStateName(in); got != want {
			t.Errorf("getStateName(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestGetRealStatusName(t *testing.T) {
	if getRealStatusName(1) != "Running" {
		t.Fail()
	}
	if getRealStatusName(99) != "Unknown" {
		t.Fail()
	}
}

func TestCleanServiceMap(t *testing.T) {
	sm := &models.ServiceMapResponse{
		Status: "ok",
		VServers: []models.VServerMap{{
			VServices: []models.VServiceMap{{
				CStatus: "OK",
				RGroup: &models.RGroupMap{
					CStatus: "OK",
					RServers: []models.RServerMap{{
						CStatus:  "OK",
						HCReason: "Not Available",
					}},
				},
			}},
		}},
	}

	cleanServiceMap(sm)

	if sm.Status != "" {
		t.Errorf("Status = %q, want empty", sm.Status)
	}
	if sm.VServers[0].VServices[0].CStatus != "" {
		t.Error("VService.CStatus no limpiado")
	}
	if sm.VServers[0].VServices[0].RGroup.CStatus != "" {
		t.Error("RGroup.CStatus no limpiado")
	}
	if sm.VServers[0].VServices[0].RGroup.RServers[0].CStatus != "" {
		t.Error("RServer.CStatus no limpiado")
	}
	if sm.VServers[0].VServices[0].RGroup.RServers[0].HCReason != "" {
		t.Error("RServer.HCReason no limpiado")
	}
}

func TestCleanServiceMap_PreservesNonOK(t *testing.T) {
	sm := &models.ServiceMapResponse{
		Status: "warning",
		VServers: []models.VServerMap{{
			VServices: []models.VServiceMap{{
				CStatus: "Failure",
				RGroup: &models.RGroupMap{
					RServers: []models.RServerMap{{
						CStatus:  "Degraded",
						HCReason: "Health check failed",
					}},
				},
			}},
		}},
	}

	cleanServiceMap(sm)

	if sm.Status != "warning" {
		t.Error("status no-ok fue limpiado")
	}
	if sm.VServers[0].VServices[0].CStatus != "Failure" {
		t.Error("cstatus no-OK fue limpiado")
	}
	if sm.VServers[0].VServices[0].RGroup.RServers[0].HCReason != "Health check failed" {
		t.Error("HCReason no 'Not Available' fue limpiado")
	}
}

func TestExtractIPFromURL(t *testing.T) {
	cases := map[string]string{
		"https://172.31.163.18":          "172.31.163.18",
		"https://172.31.163.18:8443":     "172.31.163.18",
		"https://host.local":             "host.local",
		"https://[::1]":                  "::1",
		"https://[2001:db8::1]:8443":     "2001:db8::1",
		"":                               "",
	}
	for in, want := range cases {
		if got := extractIPFromURL(in); got != want {
			t.Errorf("extractIPFromURL(%q) = %q, want %q", in, got, want)
		}
	}
}
