package commands

import (
	"bytes"
	"fmt"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "bat",
		Description:  []string{"Pretty prints the contents of a remote file.", "Syntax: 'bat <file>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdBat,
	})
}

func cmdBat(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) < 1 {
		s.Errorf("Syntax: 'bat <file>'")
		return nil
	}

	for _, arg := range args {
		remote := s.SM().Resolve(arg)

		var buf bytes.Buffer
		if _, err := s.SM().DownloadTo(remote, &buf); err != nil {
			s.Errorf("[!] SMB Error: %s", err)
			continue
		}

		if len(args) > 1 {
			s.Print(fmt.Sprintf("\x1b[1;93m[>] %s\x1b[0m", padHeader(utils.ToWirePath(remote))))
		}
		printNumbered(s, buf.String())
	}
	return nil
}
