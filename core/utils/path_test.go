package utils

import "testing"

func TestNormalizeRemotePath(t *testing.T) {
	cases := map[string]string{
		"":                "",
		"\\":              "",
		"\\\\":            "",
		"a\\b":            "a\\b",
		"\\a\\b\\":        "a\\b",
		"a/b/c":           "a\\b\\c",
		"a\\.\\b":         "a\\b",
		"a\\b\\..":        "a",
		"a\\b\\..\\..\\c": "c",
		"..\\..":          "",
	}
	for in, want := range cases {
		if got := NormalizeRemotePath(in); got != want {
			t.Errorf("NormalizeRemotePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestJoinRemotePath(t *testing.T) {
	cases := []struct {
		base, target, want string
	}{
		{"dir", "sub", "dir\\sub"},
		{"dir\\sub", "..", "dir"},
		{"dir", "\\abs\\path", "abs\\path"},
		{"", "file.txt", "file.txt"},
		{"a\\b", ".\\c", "a\\b\\c"},
	}
	for _, c := range cases {
		if got := JoinRemotePath(c.base, c.target); got != c.want {
			t.Errorf("JoinRemotePath(%q, %q) = %q, want %q", c.base, c.target, got, c.want)
		}
	}
}

func TestRemoteBase(t *testing.T) {
	cases := map[string]string{
		"":             "",
		"\\":           "",
		"file.txt":     "file.txt",
		"a\\b\\c.txt":  "c.txt",
		"\\a\\b\\":     "b",
		"dir\\sub\\..": "dir",
	}
	for in, want := range cases {
		if got := RemoteBase(in); got != want {
			t.Errorf("RemoteBase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRemoteDir(t *testing.T) {
	cases := map[string]string{
		"":            "",
		"file.txt":    "",
		"a\\b\\c.txt": "a\\b",
		"\\a\\b":      "a",
	}
	for in, want := range cases {
		if got := RemoteDir(in); got != want {
			t.Errorf("RemoteDir(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToWirePath(t *testing.T) {
	if got := ToWirePath(""); got != "\\" {
		t.Errorf("ToWirePath(\"\") = %q, want %q", got, "\\")
	}
	if got := ToWirePath("a\\b"); got != "\\a\\b" {
		t.Errorf("ToWirePath(\"a\\\\b\") = %q, want %q", got, "\\a\\b")
	}
}
