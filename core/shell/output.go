package shell

import "fmt"

// Print writes a raw line to stdout with no level prefix. Used for command
// output (listings, file contents, …).
func (s *Shell) Print(message string) {
	fmt.Println(message)
}

// Printf is Print with formatting.
func (s *Shell) Printf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Info logs an informational message as "[info] <message>".
func (s *Shell) Info(message string) {
	fmt.Printf("[\x1b[1;92minfo\x1b[0m] %s\n", message)
}

// Infof is Info with formatting.
func (s *Shell) Infof(format string, args ...interface{}) {
	s.Info(fmt.Sprintf(format, args...))
}

// Errorf logs an error as "[error] <message>".
func (s *Shell) Errorf(format string, args ...interface{}) {
	fmt.Printf("[\x1b[1;91merror\x1b[0m] %s\n", fmt.Sprintf(format, args...))
}

// Debugf logs a "[debug] <message>" line, but only when debug mode is enabled.
func (s *Shell) Debugf(format string, args ...interface{}) {
	if s.debug {
		fmt.Printf("[debug] %s\n", fmt.Sprintf(format, args...))
	}
}
