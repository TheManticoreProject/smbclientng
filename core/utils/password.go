package utils

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

// NeedsPasswordPrompt reports whether the password must be asked for
// interactively: a username was given without any secret (password, NT hash or
// Kerberos), and --no-pass was not set to request an empty password explicitly.
func NeedsPasswordPrompt(username, password, hashes string, noPass, useKerberos bool) bool {
	return username != "" && password == "" && hashes == "" && !noPass && !useKerberos
}

// ReadPassword prints prompt and reads a password without echoing it. It reads
// from the controlling terminal (/dev/tty) when there is one, so that commands
// piped on stdin are left untouched, and falls back to stdin when stdin is a
// terminal (e.g. on Windows).
func ReadPassword(prompt string) (string, error) {
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		defer tty.Close()
		fmt.Fprint(tty, prompt)
		password, err := term.ReadPassword(int(tty.Fd()))
		fmt.Fprintln(tty)
		return string(password), err
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("no terminal available to read the password from")
	}
	fmt.Fprint(os.Stderr, prompt)
	password, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	return string(password), err
}
