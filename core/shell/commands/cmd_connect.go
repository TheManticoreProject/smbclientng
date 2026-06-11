package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:        "connect",
		Description: []string{"Connect to the remote machine (useful if connection timed out).", "Syntax: 'connect'"},
		Handler:     cmdConnect,
	})
}

// cmdConnect establishes the SMB connection when it is down. If a session is
// already live it is left untouched; otherwise it connects, authenticates and
// restores the previously selected share (mirroring the Python ping_smb_session
// behaviour). Use 'reconnect' to force a fresh connection over a live one.
func cmdConnect(s *shell.Shell, args []string) error {
	if s.SM().IsConnected() {
		s.Info("Already connected. Use 'reconnect' to force a new connection.")
		return nil
	}
	if err := s.SM().Reconnect(); err != nil {
		s.Errorf("Connection failed: %s", err)
		return nil
	}
	s.Info("Connected.")
	return nil
}
