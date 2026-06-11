package shell

import (
	"fmt"
	"strings"
)

// helpColumnWidth is the width the command name column is padded to in the help
// listing (15 characters of dashes after the name).
const helpColumnWidth = 15

// PrintHelp renders the help box. With an empty command it lists every command;
// otherwise it shows the detail for a single command.
func (s *Shell) PrintHelp(command string) {
	if command != "" {
		// 'help format' explains the file-attributes column shown by ls/dir.
		if strings.ToLower(command) == "format" {
			s.printHelpFormat()
			return
		}
		cmd, ok := s.commands[strings.ToLower(command)]
		if !ok {
			s.Errorf("Help for command '%s' does not exist.", command)
			return
		}
		s.Print("│")
		s.printCommandHelpLine(cmd)
		s.Print("│")
		return
	}

	s.Print("│")
	for _, cmd := range s.order {
		s.printCommandHelpLine(cmd)
		s.Print("│")
	}
}

// printCommandHelpLine prints one command's row (and any extra description
// lines) in the help box.
func (s *Shell) printCommandHelpLine(cmd *Command) {
	dashes := helpColumnWidth - len(cmd.Name)
	if dashes < 0 {
		dashes = 0
	}
	commandStr := fmt.Sprintf("%s %s%s%s", cmd.Name, ColorGrey, strings.Repeat("─", dashes), ColorReset)

	if len(cmd.Description) == 0 {
		s.Print(fmt.Sprintf("│ ■ %s%s┤%s  ", commandStr, ColorGrey, ColorReset))
		return
	}

	s.Print(fmt.Sprintf("│ ■ %s%s┤%s %s ", commandStr, ColorGrey, ColorReset, cmd.Description[0]))
	for _, line := range cmd.Description[1:] {
		s.Print(fmt.Sprintf("│ %s%s│%s %s ", strings.Repeat(" ", helpColumnWidth+3), ColorGrey, ColorReset, line))
	}
}

// printHelpFormat explains the 8-character file-attributes string (dachnrst)
// shown in the ls/dir listings.
func (s *Shell) printHelpFormat() {
	s.Print("File attributes format:\n")
	s.Print("dachnrst")
	rows := []struct {
		tree, label string
	}{
		{"│││││││└──>", "Temporary"},
		{"││││││└───>", "System"},
		{"│││││└────>", "Read-Only"},
		{"││││└─────>", "Normal"},
		{"│││└──────>", "Hidden"},
		{"││└───────>", "Compressed"},
		{"│└────────>", "Archived"},
		{"└─────────>", "Directory"},
	}
	for _, r := range rows {
		s.Print(fmt.Sprintf("%s%s%s %s", ColorGrey, r.tree, ColorReset, r.label))
	}
}
