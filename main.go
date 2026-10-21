package main

import (
	"fmt"

	"github.com/TheManticoreProject/smbclient-ng/core/client"
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
	// Blank import registers every interactive command via the commands
	// package's init() functions.
	_ "github.com/TheManticoreProject/smbclient-ng/core/shell/commands"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/goopts/parser"
)

// version is the release shown in the banner.
const version = "1.0.0"

// banner returns the smbclient-ng ASCII art banner, with the version
// right-justified in a 10-character field. Raw-string segments keep the
// backslashes of the art literal; the two backticks on the third line are
// injected explicitly since they cannot appear inside a raw string.
func banner() string {
	art := `               _          _ _            _
 ___ _ __ ___ | |__   ___| (_) ___ _ __ | |_      _ __   __ _
/ __| '_ ` + "`" + ` _ \| '_ \ / __| | |/ _ \ '_ \| __|____| '_ \ / _` + "`" + ` |
\__ \ | | | | | |_) | (__| | |  __/ | | | ||_____| | | | (_| |
|___/_| |_| |_|_.__/ \___|_|_|\___|_| |_|\__|    |_| |_|\__, |
    by @podalirius_                         %10s  |___/`
	return fmt.Sprintf(art, "v"+version)
}

var (
	// Configuration
	debug bool

	// Target
	host   string
	port   int
	useNBT bool

	// Authentication
	authDomain   string
	authUsername string
	authPassword string
	authHashes   string
	authNoPass   bool

	// Connection Settings
	useKerberos    bool
	timeout        string
	advertisedName string
)

func parseArgs() {
	ap := parser.ArgumentsParser{
		Banner: banner(),
	}
	ap.SetOptShowBannerOnHelp(true)
	ap.SetOptShowBannerOnRun(true)

	// Configuration flags
	ap.NewBoolArgument(&debug, "", "--debug", false, "Debug mode.")

	// Target flags
	group_target, err := ap.NewArgumentGroup("Target")
	if err != nil {
		logger.Warn(fmt.Sprintf("Error creating ArgumentGroup: %s", err))
	} else {
		group_target.NewStringArgument(&host, "-H", "--host", "", true, "IP address or hostname of the SMB server to connect to.")
		group_target.NewTcpPortArgument(&port, "-P", "--port", 445, false, "Port number to connect to on the SMB server.")
		group_target.NewBoolArgument(&useNBT, "", "--nbt", false, "Use NetBIOS (NBT) transport instead of direct TCP.")
	}

	// Authentication flags
	group_auth, err := ap.NewArgumentGroup("Authentication")
	if err != nil {
		logger.Warn(fmt.Sprintf("Error creating ArgumentGroup: %s", err))
	} else {
		group_auth.NewStringArgument(&authDomain, "-d", "--domain", ".", false, "Windows domain to authenticate to. Use '.' for local accounts.")
		group_auth.NewStringArgument(&authUsername, "-u", "--user", "", false, "User to authenticate as. Supports DOMAIN\\user.")
		group_auth.NewStringArgument(&authPassword, "-p", "--password", "", false, "Password to authenticate with. Prompted for when no secret is given.")
		group_auth.NewBoolArgument(&authNoPass, "", "--no-pass", false, "Do not prompt for a password (authenticate with an empty one).")
		group_auth.NewStringArgument(&authHashes, "", "--hashes", "", false, "NT/LM hashes for pass-the-hash, format is [LMhash:]NThash.")
	}

	// Connection Settings flags
	group_conn, err := ap.NewArgumentGroup("Connection Settings")
	if err != nil {
		logger.Warn(fmt.Sprintf("Error creating ArgumentGroup: %s", err))
	} else {
		group_conn.NewBoolArgument(&useKerberos, "-k", "--use-kerberos", false, "Use Kerberos instead of NTLM. (not yet supported by the Manticore SMB stack)")
		group_conn.NewStringArgument(&timeout, "-T", "--timeout", "3", false, "Timeout in seconds of each connection attempt (TCP connect and dialect negotiation).")
		group_conn.NewStringArgument(&advertisedName, "-a", "--advertised-name", "", false, "Machine name advertised to the server during authentication.")
	}

	ap.Parse()
}

func main() {
	parseArgs()

	connectTimeout, err := utils.ParseTimeout(timeout)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// A username is required. Guard against an empty one up front: otherwise the SMB
	// stack attempts a null/anonymous session and the server returns an opaque
	// LOGON_FAILURE.
	if authUsername == "" {
		logger.Error("A username is required: provide one with -u/--user.")
		return
	}

	// Accept DOMAIN\user and DOMAIN/user; the embedded domain overrides -d.
	authDomain, authUsername = utils.SplitDomainUsername(authDomain, authUsername)

	if authNoPass && (authPassword != "" || authHashes != "") {
		logger.Error("--no-pass cannot be combined with -p/--password or --hashes.")
		return
	}
	if utils.NeedsPasswordPrompt(authUsername, authPassword, authHashes, authNoPass, useKerberos) {
		password, err := utils.ReadPassword(fmt.Sprintf("  | Provide a password for '%s\\%s': ", authDomain, authUsername))
		if err != nil {
			logger.Error(fmt.Sprintf("Could not read the password: %s (use -p, --hashes or --no-pass)", err))
			return
		}
		authPassword = password
	}

	creds, err := credentials.NewCredentials(authDomain, authUsername, authPassword, authHashes)
	if err != nil {
		logger.Warn(fmt.Sprintf("Error creating credentials: %s", err))
		return
	}

	// Build the session manager wrapping the Manticore SMB v1.0 client.
	sm := client.NewSessionManager(host, port, useNBT, creds, useKerberos, debug)
	sm.Timeout = connectTimeout
	sm.AdvertisedName = advertisedName

	// Establish the transport connection and authenticate before dropping into
	// the interactive shell.
	if err := sm.Connect(); err != nil {
		logger.Error(fmt.Sprintf("Could not connect to %s:%d: %s", host, port, err))
		return
	}

	if err := sm.Login(); err != nil {
		logger.Error(fmt.Sprintf("Authentication failed: %s", err))
		sm.Close()
		return
	}

	logger.Info(fmt.Sprintf("Connected and authenticated to %s:%d.", host, port))

	// Launch the interactive REPL.
	sh := shell.NewShell(sm, debug)
	sh.Run()
}
