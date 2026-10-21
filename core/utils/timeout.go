package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ParseTimeout parses a timeout given in seconds, possibly fractional (e.g.
// "3" or "0.5"), as accepted by the -T/--timeout flag. The timeout must be
// strictly positive.
func ParseTimeout(seconds string) (time.Duration, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(seconds), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("invalid timeout %q: expected a number of seconds", seconds)
	}
	if value <= 0 {
		return 0, fmt.Errorf("invalid timeout %q: must be greater than zero", seconds)
	}
	timeout := time.Duration(value * float64(time.Second))
	if timeout <= 0 {
		return 0, fmt.Errorf("invalid timeout %q: too small", seconds)
	}
	return timeout, nil
}
