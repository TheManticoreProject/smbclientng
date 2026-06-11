package commands

import (
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "mkdir",
		Description:  []string{"Creates a new remote directory.", "Syntax: 'mkdir <directory>'"},
		Autocomplete: []string{"remote_directory"},
		Handler:      cmdMkdir,
	})
}

func cmdMkdir(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) != 1 {
		s.Errorf("Syntax: 'mkdir <directory>'")
		return nil
	}
	target := s.SM().Resolve(args[0])
	if err := s.SM().MakeDir(target); err != nil {
		s.Errorf("Error creating directory '%s': %s", utils.ToWirePath(target), err)
		return nil
	}
	return nil
}
