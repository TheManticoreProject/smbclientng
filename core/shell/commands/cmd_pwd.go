package commands

import (
	"fmt"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:        "pwd",
		Description: []string{"Print the current remote working directory.", "Syntax: 'pwd'"},
		Handler:     cmdPwd,
	})
}

func cmdPwd(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	cwd := s.SM().Cwd()
	if cwd != "" {
		cwd += "\\"
	}
	s.Print(fmt.Sprintf("\\\\%s\\%s\\%s", s.SM().Host, s.SM().CurrentShare(), cwd))
	return nil
}
