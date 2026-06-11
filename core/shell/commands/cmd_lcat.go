package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "lcat",
		Description:  []string{"Print the contents of a local file.", "Syntax: 'lcat <file>'"},
		Autocomplete: []string{"local_file"},
		Handler:      cmdLcat,
	})
}

func cmdLcat(s *shell.Shell, args []string) error {
	if len(args) < 1 {
		s.Errorf("Syntax: 'lcat <file>'")
		return nil
	}

	for _, arg := range args {
		target := localPath(s, arg)
		content, err := os.ReadFile(target)
		if err != nil {
			s.Errorf("[!] Local file '%s' does not exist.", arg)
			continue
		}
		if len(args) > 1 {
			s.Print(fmt.Sprintf("\x1b[1;93m[>] %s\x1b[0m", padHeader(arg)))
		}
		s.Print(strings.TrimRight(string(content), "\r\n"))
	}
	return nil
}

// padHeader renders the "<path> ====...===" banner used by the multi-file
// cat/head/tail family, padded with '=' to 80 columns.
func padHeader(path string) string {
	header := path + " "
	if len(header) < 80 {
		header += strings.Repeat("=", 80-len(header))
	}
	return header
}
