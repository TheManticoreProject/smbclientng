package commands

import "testing"

// TestFormatInfoRow checks the info-tree row formatting (colors and dashes).
func TestFormatInfoRow(t *testing.T) {
	cases := []struct {
		lead   string
		label  string
		dashes int
		value  string
		want   string
	}{
		{"  │ ├─ ", "NetBIOS Hostname", 8, "DC01",
			"  │ ├─ \x1b[94mNetBIOS Hostname\x1b[0m \x1b[90m────────\x1b[0m : \x1b[93mDC01\x1b[0m"},
		{"  │ └─ ", "DNS Domain", 14, "corp.local",
			"  │ └─ \x1b[94mDNS Domain\x1b[0m \x1b[90m──────────────\x1b[0m : \x1b[93mcorp.local\x1b[0m"},
		{"  │ └─ ", "Max size of write chunk", 1, "1048576 bytes (1.00 MB)",
			"  │ └─ \x1b[94mMax size of write chunk\x1b[0m \x1b[90m─\x1b[0m : \x1b[93m1048576 bytes (1.00 MB)\x1b[0m"},
		{"  ├─ ", "Description", 5, "Remote IPC",
			"  ├─ \x1b[94mDescription\x1b[0m \x1b[90m─────\x1b[0m : \x1b[93mRemote IPC\x1b[0m"},
		{"  │ ├─ ", "OS Name", 17, "Windows Server 2019",
			"  │ ├─ \x1b[94mOS Name\x1b[0m \x1b[90m─────────────────\x1b[0m : \x1b[93mWindows Server 2019\x1b[0m"},
		{"  │ ├─ ", "Supports NTLMv2", 9, "True",
			"  │ ├─ \x1b[94mSupports NTLMv2\x1b[0m \x1b[90m─────────\x1b[0m : \x1b[93mTrue\x1b[0m"},
	}
	for _, c := range cases {
		if got := formatInfoRow(c.lead, c.label, c.dashes, c.value); got != c.want {
			t.Errorf("formatInfoRow(%q,%d) =\n  %q\nwant\n  %q", c.label, c.dashes, got, c.want)
		}
	}
}

func TestPyBool(t *testing.T) {
	if pyBool(true) != "True" || pyBool(false) != "False" {
		t.Errorf("pyBool mismatch: %q %q", pyBool(true), pyBool(false))
	}
}
