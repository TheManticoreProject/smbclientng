package shell

import (
	"reflect"
	"testing"
)

func TestLastWord(t *testing.T) {
	cases := map[string]string{
		"use c":       "c",
		"use ":        "",
		"use":         "use",
		"use  C$":     "C$",
		"info --shar": "--shar",
	}
	for in, want := range cases {
		if got := lastWord(in); got != want {
			t.Errorf("lastWord(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCaseInsensitiveComplete(t *testing.T) {
	shares := []string{"C$", "ADMIN$", "IPC$", "Users"}

	cases := []struct {
		partial string
		want    []string
	}{
		// Lowercase typing matches case-insensitively; the typed casing is kept
		// for the matched prefix (readline cannot rewrite typed characters).
		{"c", []string{"c$"}},
		// Exact-case typing completes to the real share name.
		{"C", []string{"C$"}},
		{"ad", []string{"adMIN$"}},
		{"ADMIN$", []string{"ADMIN$"}},
		// Empty partial lists every share in its real case.
		{"", []string{"ADMIN$", "C$", "IPC$", "Users"}},
		// No case-insensitive match.
		{"x", []string{}},
		// 'i' matches IPC$ only.
		{"i", []string{"iPC$"}},
	}
	for _, c := range cases {
		got := caseInsensitiveComplete(c.partial, shares)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("caseInsensitiveComplete(%q) = %#v, want %#v", c.partial, got, c.want)
		}
	}
}
