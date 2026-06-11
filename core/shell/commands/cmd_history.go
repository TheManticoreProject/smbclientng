package commands

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name: "history",
		Description: []string{
			"Displays the command history.",
			"Syntax: 'history [--contains <string>] [--clear]'",
		},
		Handler: cmdHistory,
	})
}

func cmdHistory(s *shell.Shell, args []string) error {
	var contains string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--clear":
			s.ClearHistory()
			s.Info("Command history cleared.")
			return nil
		case "--contains":
			if i+1 >= len(args) {
				s.Errorf("Syntax: 'history [--contains <string>] [--clear]'")
				return nil
			}
			contains = args[i+1]
			i++
		default:
			s.Errorf("Unknown argument '%s'. Syntax: 'history [--contains <string>] [--clear]'", args[i])
			return nil
		}
	}

	entries := s.History()
	// Width of the index column, right-aligned to the highest line number.
	width := len(fmt.Sprintf("%d", len(entries)))
	for i, line := range entries {
		if contains != "" && !strings.Contains(line, contains) {
			continue
		}
		s.Print(fmt.Sprintf("%*d | %s", width, i+1, line))
	}
	return nil
}
