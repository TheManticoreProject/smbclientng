package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:        "lpwd",
		Description: []string{"Shows the current local directory.", "Syntax: 'lpwd'"},
		Handler:     cmdLpwd,
	})
}

func cmdLpwd(s *shell.Shell, args []string) error {
	s.Print(s.LocalCwd())
	return nil
}
