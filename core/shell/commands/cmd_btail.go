package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "btail",
		Description:  []string{"Pretty prints the last <n> lines of a remote file.", "Syntax: 'btail [-n <lines>] <file>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdBtail,
	})
}

func cmdBtail(s *shell.Shell, args []string) error {
	return headTail(s, args, true, true)
}
