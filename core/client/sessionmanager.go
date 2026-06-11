package client

import (
	"fmt"

	"github.com/TheManticoreProject/smbclient-ng/core/utils"

	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// SessionManager wraps the Manticore generic SMB client and tracks the
// interactive shell's connection state: the current share and the current
// working directory within that share. It is the single point through which the
// shell commands talk to the server.
//
// The generic client negotiates the dialect for us (preferring the highest the
// server supports), so the shell speaks SMB1, SMB2, or SMB3 transparently.
type SessionManager struct {
	Host        string
	Port        int
	UseNBT      bool
	Creds       *credentials.Credentials
	UseKerberos bool
	Debug       bool

	client *smbclient.Client

	// share is the name of the currently selected share ("" when none is
	// selected). cwd is the normalized, share-relative working directory
	// (backslash separated, no leading/trailing separator; "" is the root).
	share string
	cwd   string
}

// NewSessionManager creates a SessionManager. It does not open any connection;
// call Connect followed by Login.
func NewSessionManager(host string, port int, useNBT bool, creds *credentials.Credentials, useKerberos bool, debug bool) *SessionManager {
	return &SessionManager{
		Host:        host,
		Port:        port,
		UseNBT:      useNBT,
		Creds:       creds,
		UseKerberos: useKerberos,
		Debug:       debug,
	}
}

// Connect resolves the target, dials it with the generic SMB client, and
// negotiates the highest dialect the server and client share.
func (sm *SessionManager) Connect() error {
	if sm.UseNBT {
		return fmt.Errorf("NBT transport is not supported by the generic SMB client; use direct TCP (the default)")
	}

	// Empty Options: offer all supported dialects, best first (SMB2/SMB3 before
	// SMB1), and let the server pick the highest it supports.
	c, err := smbclient.Dial(sm.Host, sm.Port, smbclient.Options{})
	if err != nil {
		return err
	}
	sm.client = c
	return nil
}

// Login authenticates the session. Kerberos and pass-the-hash are not yet wired
// through the Manticore SMB stack, so we surface a clear error rather than
// silently doing the wrong thing.
func (sm *SessionManager) Login() error {
	if sm.client == nil {
		return fmt.Errorf("not connected")
	}
	if sm.UseKerberos {
		return fmt.Errorf("Kerberos authentication is not yet supported by the Manticore SMB stack")
	}
	if sm.Creds != nil && sm.Creds.GetPassword() == "" && sm.Creds.CanPassTheHash() {
		return fmt.Errorf("pass-the-hash is not yet supported by the Manticore SMB stack; provide a password instead")
	}
	return sm.client.Login(sm.Creds)
}

// Dialect returns a human-readable name of the negotiated SMB dialect.
func (sm *SessionManager) Dialect() string {
	if sm.client == nil {
		return "(not connected)"
	}
	return sm.client.Dialect().String()
}

// UseShare connects to the named share (tree connect) and resets the working
// directory to the share root.
func (sm *SessionManager) UseShare(name string) error {
	if sm.client == nil {
		return fmt.Errorf("not connected")
	}
	if err := sm.client.TreeConnect(name); err != nil {
		return err
	}
	sm.share = name
	sm.cwd = ""
	return nil
}

// CurrentShare returns the selected share name ("" if none).
func (sm *SessionManager) CurrentShare() string {
	return sm.share
}

// Cwd returns the normalized share-relative working directory.
func (sm *SessionManager) Cwd() string {
	return sm.cwd
}

// PromptPath returns a human-friendly path for the shell prompt, e.g.
// "\\" when at a share root or "\\dir\\sub" deeper in the tree.
func (sm *SessionManager) PromptPath() string {
	return utils.ToWirePath(sm.cwd)
}

// Resolve turns a user-supplied path argument (relative to the cwd or absolute
// to the share root) into a normalized share-relative path.
func (sm *SessionManager) Resolve(arg string) string {
	return utils.JoinRemotePath(sm.cwd, arg)
}

// ChangeDir verifies that the resolved target directory exists and, if so,
// updates the working directory.
func (sm *SessionManager) ChangeDir(arg string) error {
	if sm.share == "" {
		return fmt.Errorf("no share selected; use 'use <share>' first")
	}
	target := sm.Resolve(arg)
	// The share root always exists and is not openable as a named directory on
	// every dialect, so only verify non-root targets.
	if target != "" {
		if err := sm.client.CheckDirectory(target); err != nil {
			return fmt.Errorf("cannot change directory to %q: %w", utils.ToWirePath(target), err)
		}
	}
	sm.cwd = target
	return nil
}

// List enumerates the entries of the resolved directory (or the cwd when arg is
// empty).
func (sm *SessionManager) List(arg string) ([]smbclient.FileInfo, error) {
	if sm.share == "" {
		return nil, fmt.Errorf("no share selected; use 'use <share>' first")
	}
	target := sm.cwd
	if arg != "" {
		target = sm.Resolve(arg)
	}
	return sm.client.ListDirectory(target, "*")
}

// Client exposes the underlying generic SMB client for commands that need direct
// access (recursive listing, etc.).
func (sm *SessionManager) Client() *smbclient.Client {
	return sm.client
}

// IsConnected reports whether a client has been instantiated.
func (sm *SessionManager) IsConnected() bool {
	return sm.client != nil
}

// Reconnect tears down the current client and re-establishes the connection,
// authentication and (if one was selected) the share.
func (sm *SessionManager) Reconnect() error {
	share := sm.share
	cwd := sm.cwd

	sm.Close()

	if err := sm.Connect(); err != nil {
		return err
	}
	if err := sm.Login(); err != nil {
		return err
	}
	if share != "" {
		if err := sm.UseShare(share); err != nil {
			return err
		}
		sm.cwd = cwd
	}
	return nil
}

// Close performs a best-effort clean teardown of the session: tree disconnect,
// logoff, then transport disconnect. Errors are ignored since this runs during
// shutdown.
func (sm *SessionManager) Close() {
	if sm.client == nil {
		return
	}
	if sm.share != "" {
		_ = sm.client.TreeDisconnect()
	}
	_ = sm.client.Logoff()
	_ = sm.client.Disconnect()
	sm.client = nil
	sm.share = ""
	sm.cwd = ""
}
