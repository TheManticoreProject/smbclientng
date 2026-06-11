package shell

// ANSI color codes used across the shell output. They are exported so the
// commands package can reuse the exact same palette.
const (
	ColorReset    = "\x1b[0m"
	ColorBold     = "\x1b[1m"
	ColorBoldCyan = "\x1b[1;96m"
	ColorGrey     = "\x1b[90m"
)
