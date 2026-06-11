package commands

import (
	"bytes"
	"fmt"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "head",
		Description:  []string{"Get the first <n> lines of a remote file.", "Syntax: 'head [-n <lines>] <file>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdHead,
	})
}

func cmdHead(s *shell.Shell, args []string) error {
	return headTail(s, args, false, false)
}

// headTail implements the head/tail/bhead/btail family: it downloads each
// remote file, keeps the first (or last) n lines, and prints it plainly or with
// a line-number gutter when pretty is set.
func headTail(s *shell.Shell, args []string, tail, pretty bool) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}

	n, files := parseLinesFlag(args)
	if len(files) < 1 {
		s.Errorf("Syntax: requires at least one <file>")
		return nil
	}

	for _, arg := range files {
		remote := s.SM().Resolve(arg)

		var buf bytes.Buffer
		if _, err := s.SM().DownloadTo(remote, &buf); err != nil {
			s.Errorf("[!] SMB Error: %s", err)
			continue
		}

		if len(files) > 1 {
			s.Print(fmt.Sprintf("\x1b[1;93m[>] %s\x1b[0m", padHeader(utils.ToWirePath(remote))))
		}

		content := limitLines(buf.String(), n, tail)
		if pretty {
			printNumbered(s, content)
		} else {
			s.Print(content)
		}
	}
	return nil
}
