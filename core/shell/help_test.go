package shell

import (
	"io"
	"os"
	"testing"

	"github.com/TheManticoreProject/smbclient-ng/core/client"
)

// captureStdout runs f with os.Stdout redirected to a pipe and returns what it
// wrote.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

// TestPrintHelpFormat checks the `help format` file-attributes legend output.
func TestPrintHelpFormat(t *testing.T) {
	s := NewShell(client.NewSessionManager("host", 445, false, nil, false, false), false)
	got := captureStdout(t, func() { s.PrintHelp("format") })

	want := "File attributes format:\n\n" +
		"dachnrst\n" +
		"\x1b[90m│││││││└──>\x1b[0m Temporary\n" +
		"\x1b[90m││││││└───>\x1b[0m System\n" +
		"\x1b[90m│││││└────>\x1b[0m Read-Only\n" +
		"\x1b[90m││││└─────>\x1b[0m Normal\n" +
		"\x1b[90m│││└──────>\x1b[0m Hidden\n" +
		"\x1b[90m││└───────>\x1b[0m Compressed\n" +
		"\x1b[90m│└────────>\x1b[0m Archived\n" +
		"\x1b[90m└─────────>\x1b[0m Directory\n"

	if got != want {
		t.Errorf("PrintHelp(\"format\") =\n%q\nwant\n%q", got, want)
	}
}
