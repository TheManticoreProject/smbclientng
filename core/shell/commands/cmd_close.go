package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:        "close",
		Description: []string{"Closes the SMB connection to the remote machine.", "Syntax: 'close'"},
		Handler:     cmdClose,
	})
}

func cmdClose(s *shell.Shell, args []string) error {
	s.SM().Close()
	s.Info("Session closed. Use 'reconnect' to open a new one.")
	return nil
}
