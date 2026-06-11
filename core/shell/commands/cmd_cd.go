package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "cd",
		Description:  []string{"Change the current working directory.", "Syntax: 'cd <directory>'"},
		Autocomplete: []string{"remote_directory"},
		Handler:      cmdCd,
	})
}

func cmdCd(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	target := ""
	if len(args) >= 1 {
		target = args[0]
	}
	if err := s.SM().ChangeDir(target); err != nil {
		s.Errorf("[!] SMB Error: %s", err)
	}
	return nil
}
