package commands

import (
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "mv",
		Aliases:      []string{"rename", "move"},
		Description:  []string{"Move or rename a remote file or directory.", "Syntax: 'mv <source> <destination>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdMv,
	})
}

func cmdMv(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) != 2 {
		s.Errorf("Syntax: 'mv <source> <destination>'")
		return nil
	}
	src := s.SM().Resolve(args[0])
	dst := s.SM().Resolve(args[1])
	if err := s.SM().Rename(src, dst); err != nil {
		s.Errorf("[!] SMB Error: %s", err)
		return nil
	}
	s.Infof("Renamed '%s' to '%s'.", utils.ToWirePath(src), utils.ToWirePath(dst))
	return nil
}
