package utils

import "testing"

func TestNeedsPasswordPrompt(t *testing.T) {
	cases := []struct {
		name                     string
		username, password, hash string
		noPass, kerberos         bool
		want                     bool
	}{
		{"username only", "user", "", "", false, false, true},
		{"password given", "user", "secret", "", false, false, false},
		{"hash given", "user", "", ":fc525c9683e8fe067095ba2ddc971889", false, false, false},
		{"no-pass", "user", "", "", true, false, false},
		{"kerberos", "user", "", "", false, true, false},
		{"no username", "", "", "", false, false, false},
	}
	for _, c := range cases {
		if got := NeedsPasswordPrompt(c.username, c.password, c.hash, c.noPass, c.kerberos); got != c.want {
			t.Errorf("%s: NeedsPasswordPrompt() = %v, want %v", c.name, got, c.want)
		}
	}
}
