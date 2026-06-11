package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "tail",
		Description:  []string{"Get the last <n> lines of a remote file.", "Syntax: 'tail [-n <lines>] <file>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdTail,
	})
}

func cmdTail(s *shell.Shell, args []string) error {
	return headTail(s, args, true, false)
}
