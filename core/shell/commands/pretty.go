package commands

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

// printNumbered renders file contents with right-aligned line numbers in a
// grey gutter. Go has no built-in syntax highlighter, so this provides the line-number gutter
// without language-aware coloring.
func printNumbered(s *shell.Shell, content string) {
	content = strings.TrimRight(content, " \t\r\n")
	lines := strings.Split(content, "\n")
	width := len(fmt.Sprintf("%d", len(lines)))
	for i, line := range lines {
		s.Print(fmt.Sprintf("%s%*d%s │ %s", shell.ColorGrey, width, i+1, shell.ColorReset, line))
	}
}

// limitLines returns the first n (head) or last n (tail) lines of content. When
// n <= 0 or content has fewer than n lines, the whole content is returned.
func limitLines(content string, n int, tail bool) string {
	content = strings.TrimRight(content, " \t\r\n")
	if n <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= n {
		return content
	}
	if tail {
		lines = lines[len(lines)-n:]
	} else {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

// parseLinesFlag scans args for a "-n <count>" / "--lines <count>" option and
// returns the requested line count (default 10) along with the remaining
// positional arguments.
func parseLinesFlag(args []string) (int, []string) {
	n := 10
	var rest []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n", "--lines":
			if i+1 < len(args) {
				// -n always consumes its value; keep the default if it does
				// not parse as an integer.
				var v int
				if c, err := fmt.Sscanf(args[i+1], "%d", &v); err == nil && c == 1 {
					n = v
				}
				i++
			}
		default:
			rest = append(rest, args[i])
		}
	}
	return n, rest
}
