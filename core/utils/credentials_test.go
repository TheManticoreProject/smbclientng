package utils

import "testing"

func TestSplitDomainUsername(t *testing.T) {
	cases := []struct {
		domain, username     string
		wantDomain, wantUser string
	}{
		{".", "Administrator", ".", "Administrator"},
		{".", "LAB\\Administrator", "LAB", "Administrator"},
		{".", "LAB/Administrator", "LAB", "Administrator"},
		{"OTHER", "LAB\\Administrator", "LAB", "Administrator"},
		{"LAB", "Administrator", "LAB", "Administrator"},
		{".", "LAB\\sub\\user", "LAB", "sub\\user"},
		{".", "\\Administrator", ".", "\\Administrator"},
		{".", "LAB\\", ".", "LAB\\"},
		{".", "", ".", ""},
		{".", "user@lab.local", ".", "user@lab.local"},
	}
	for _, c := range cases {
		gotDomain, gotUser := SplitDomainUsername(c.domain, c.username)
		if gotDomain != c.wantDomain || gotUser != c.wantUser {
			t.Errorf("SplitDomainUsername(%q, %q) = (%q, %q), want (%q, %q)",
				c.domain, c.username, gotDomain, gotUser, c.wantDomain, c.wantUser)
		}
	}
}
