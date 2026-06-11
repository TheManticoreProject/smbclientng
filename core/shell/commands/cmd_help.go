package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "help",
		Description:  []string{"Displays this help message.", "Syntax: 'help [command]'"},
		Autocomplete: []string{"command"},
		Handler:      cmdHelp,
	})
}

func cmdHelp(s *shell.Shell, args []string) error {
	command := ""
	if len(args) >= 1 {
		command = args[0]
	}
	s.PrintHelp(command)
	return nil
}
