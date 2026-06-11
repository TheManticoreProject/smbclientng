package commands

import (
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "rmdir",
		Description:  []string{"Removes a remote directory.", "Syntax: 'rmdir <directory>'"},
		Autocomplete: []string{"remote_directory"},
		Handler:      cmdRmdir,
	})
}

func cmdRmdir(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) != 1 {
		s.Errorf("Syntax: 'rmdir <directory>'")
		return nil
	}
	target := s.SM().Resolve(args[0])
	if err := s.SM().RemoveDir(target); err != nil {
		s.Errorf("Error removing directory '%s': %s", utils.ToWirePath(target), err)
		return nil
	}
	return nil
}
