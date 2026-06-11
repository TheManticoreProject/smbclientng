package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:        "exit",
		Aliases:     []string{"quit"},
		Description: []string{"Exits the smbclient-ng script.", "Syntax: 'exit'"},
		Handler:     cmdExit,
	})
}

func cmdExit(s *shell.Shell, args []string) error {
	s.Stop()
	return nil
}
