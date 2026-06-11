package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "use",
		Description:  []string{"Use a SMB share.", "Syntax: 'use <sharename>'"},
		Autocomplete: []string{"share"},
		Handler:      cmdUse,
	})
}

func cmdUse(s *shell.Shell, args []string) error {
	if len(args) != 1 {
		s.Errorf("Syntax: 'use <sharename>'")
		return nil
	}
	if err := s.SM().UseShare(args[0]); err != nil {
		s.Errorf("No share named '%s' on '%s'", args[0], s.SM().Host)
		return nil
	}
	return nil
}
