package utils

import (
	"testing"
	"time"
)

func TestParseTimeout(t *testing.T) {
	valid := map[string]time.Duration{
		"3":    3 * time.Second,
		"0.5":  500 * time.Millisecond,
		" 10 ": 10 * time.Second,
		"1.25": 1250 * time.Millisecond,
	}
	for in, want := range valid {
		got, err := ParseTimeout(in)
		if err != nil || got != want {
			t.Errorf("ParseTimeout(%q) = (%v, %v), want (%v, nil)", in, got, err, want)
		}
	}

	for _, in := range []string{"", "abc", "0", "-1", "NaN", "Inf", "1e-12"} {
		if got, err := ParseTimeout(in); err == nil {
			t.Errorf("ParseTimeout(%q) = %v, want an error", in, got)
		}
	}
}
