package commands

import (
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "rm",
		Aliases:      []string{"del"},
		Description:  []string{"Removes a remote file.", "Syntax: 'rm <file>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdRm,
	})
}

func cmdRm(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) != 1 {
		s.Errorf("Syntax: 'rm <file>'")
		return nil
	}
	target := s.SM().Resolve(args[0])
	if err := s.SM().Delete(target); err != nil {
		s.Errorf("Error removing file '%s': %s", utils.ToWirePath(target), err)
		return nil
	}
	return nil
}
