package commands

import (
	"os"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:        "reset",
		Description: []string{"Reset the TTY output, useful if it was broken after printing a binary file on stdout.", "Syntax: 'reset'"},
		Handler:     cmdReset,
	})
}

func cmdReset(s *shell.Shell, args []string) error {
	// Show the cursor and reset terminal rendition.
	os.Stdout.WriteString("\x1b[?25h") // cursor on
	os.Stdout.WriteString("\x1b[v")
	os.Stdout.WriteString("\x1b[o") // reset
	return nil
}
