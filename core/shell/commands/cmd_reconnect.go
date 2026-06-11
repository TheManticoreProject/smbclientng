package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:        "reconnect",
		Description: []string{"Reconnect to the remote machine (useful if connection timed out).", "Syntax: 'reconnect'"},
		Handler:     cmdReconnect,
	})
}

func cmdReconnect(s *shell.Shell, args []string) error {
	if err := s.SM().Reconnect(); err != nil {
		s.Errorf("Reconnect failed: %s", err)
		return nil
	}
	s.Info("Reconnected.")
	return nil
}
