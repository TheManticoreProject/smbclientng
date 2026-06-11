package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "bhead",
		Description:  []string{"Pretty prints the first <n> lines of a remote file.", "Syntax: 'bhead [-n <lines>] <file>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdBhead,
	})
}

func cmdBhead(s *shell.Shell, args []string) error {
	return headTail(s, args, false, true)
}
