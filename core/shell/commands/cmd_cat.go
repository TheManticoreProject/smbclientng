package commands

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "cat",
		Description:  []string{"Get the contents of a remote file.", "Syntax: 'cat <file>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdCat,
	})
}

func cmdCat(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) < 1 {
		s.Errorf("Syntax: 'cat <file>'")
		return nil
	}

	for _, arg := range args {
		remote := s.SM().Resolve(arg)

		// Read the whole file, then print the multi-file header (only on success)
		// followed by the right-stripped contents.
		var buf bytes.Buffer
		if _, err := s.SM().DownloadTo(remote, &buf); err != nil {
			s.Errorf("[!] SMB Error: %s", err)
			continue
		}

		if len(args) > 1 {
			s.Print(fmt.Sprintf("\x1b[1;93m[>] %s\x1b[0m", padHeader(utils.ToWirePath(remote))))
		}
		s.Print(strings.TrimRight(buf.String(), " \t\r\n"))
	}
	return nil
}
