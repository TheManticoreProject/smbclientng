package commands

import "github.com/TheManticoreProject/smbclient-ng/core/shell"

// This file registers placeholder commands for features that the
// Manticore SMB stack does not expose yet (DCE/RPC srvsvc/samr, snapshots,
// security descriptors, …). Each prints a clear "not yet supported" message
// stating what it needs. They are grouped here because none has a real
// implementation; once Manticore grows the capability, each should graduate to
// its own cmd_<name>.go file with a full implementation.

// unsupportedCommand describes a placeholder command and why it is unavailable.
type unsupportedCommand struct {
	name    string
	summary string
	reason  string
}

var unsupportedCommands = []unsupportedCommand{
	{"sessions", "List the active sessions on the server.", "requires DCE/RPC srvsvc (NetrSessionEnum)"},
	{"mount", "Creates a mount point of the remote share on the local machine.", "requires OS-level SMB mounting"},
	{"umount", "Removes a mount point of the remote share on the local machine.", "requires OS-level SMB mounting"},
	{"acls", "List ACLs of files and folders in the current directory.", "requires SMB security-descriptor queries"},
}

func init() {
	for _, uc := range unsupportedCommands {
		uc := uc
		shell.RegisterCommand(&shell.Command{
			Name:        uc.name,
			Description: []string{uc.summary, "Not yet supported: " + uc.reason + "."},
			Handler: func(s *shell.Shell, args []string) error {
				s.Errorf("'%s' is not yet supported: %s.", uc.name, uc.reason)
				return nil
			},
		})
	}
}
