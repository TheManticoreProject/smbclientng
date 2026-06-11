package shell

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/client"

	"github.com/chzyer/readline"
)

// Shell is the interactive REPL that drives an SMB SessionManager, on top of the
// Manticore SMB stack, using github.com/chzyer/readline for line editing, command
// history and TAB completion.
type Shell struct {
	sm    *client.SessionManager
	debug bool

	localCwd string
	running  bool

	// history records the command lines entered this session, in order, for the
	// `history` command. readline keeps its own persistent history for line
	// editing; this in-memory copy is what `history` reads and can clear.
	history []string

	// commands maps a command name (and each of its aliases) to its definition.
	commands map[string]*Command
	// order preserves the canonical command list (sorted on registration) used
	// by help and the completer.
	order []*Command
}

// Command describes a single interactive command. Description holds the help
// lines shown by `help` (the first line is the short summary); Autocomplete
// lists the completion hints understood by the completer ("remote_directory",
// "remote_file", "local_directory", "share", "command").
type Command struct {
	Name         string
	Aliases      []string
	Description  []string
	Autocomplete []string
	Handler      func(s *Shell, args []string) error
}

// commandRegistry collects every command registered by the per-command files'
// init() functions. NewShell snapshots it into each shell instance.
var commandRegistry []*Command

// RegisterCommand adds a command to the package-level registry. It is called
// from the init() of each command file in the commands package so that every
// command is self-contained in its own file.
func RegisterCommand(cmd *Command) {
	commandRegistry = append(commandRegistry, cmd)
}

// SM returns the session manager driving this shell.
func (s *Shell) SM() *client.SessionManager { return s.sm }

// LocalCwd returns the current local working directory.
func (s *Shell) LocalCwd() string { return s.localCwd }

// SetLocalCwd updates the current local working directory.
func (s *Shell) SetLocalCwd(dir string) { s.localCwd = dir }

// Stop requests the read-eval-print loop to exit after the current command.
func (s *Shell) Stop() { s.running = false }

// History returns the command lines entered this session, in order.
func (s *Shell) History() []string { return s.history }

// ClearHistory empties the in-memory command history.
func (s *Shell) ClearHistory() { s.history = nil }

// NewShell builds a Shell bound to the given session manager and indexes all
// registered commands.
func NewShell(sm *client.SessionManager, debug bool) *Shell {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	s := &Shell{
		sm:       sm,
		debug:    debug,
		localCwd: cwd,
		commands: map[string]*Command{},
	}

	// Index the registry, sorted by name for stable help/completion ordering.
	s.order = append(s.order, commandRegistry...)
	sort.Slice(s.order, func(i, j int) bool { return s.order[i].Name < s.order[j].Name })
	for _, cmd := range s.order {
		s.commands[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			s.commands[alias] = cmd
		}
	}
	return s
}

// Run starts the read-eval-print loop. It returns when the user exits or stdin
// is closed. Command history (with up/down navigation) is persisted across
// sessions and TAB triggers completion.
func (s *Shell) Run() {
	s.running = true

	historyFile := filepath.Join(os.TempDir(), ".smbclient-ng_history")
	rl, err := readline.NewEx(&readline.Config{
		Prompt:            s.prompt(),
		HistoryFile:       historyFile,
		AutoComplete:      s.newCompleter(),
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		s.Errorf("Error initializing interactive shell: %s", err)
		return
	}
	defer rl.Close()

	for s.running {
		rl.SetPrompt(s.prompt())

		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			// Ctrl-C clears a non-empty line; on an empty line it exits.
			if len(line) == 0 {
				break
			}
			continue
		} else if err == io.EOF {
			// Ctrl-D closes the shell.
			break
		}

		tokens := tokenize(line)
		if len(tokens) == 0 {
			continue
		}

		s.history = append(s.history, strings.TrimRight(line, "\r\n"))

		name := strings.ToLower(tokens[0])
		cmd, ok := s.commands[name]
		if !ok {
			s.Errorf("Unknown command. Type \"help\" for help.")
			continue
		}

		if err := cmd.Handler(s, tokens[1:]); err != nil {
			s.Errorf("%s", err)
		}
	}

	s.sm.Close()
}

// prompt renders the interactive prompt: a colored connection dot, then the
// current UNC path in blue. With no share selected the
// path is "\\<host>\"; once a share is in use it is "\\<host>\<share>\<cwd>\".
func (s *Shell) prompt() string {
	var dot string
	if s.sm.IsConnected() {
		dot = "\x1b[1;92m■\x1b[0m" // green ■
	} else {
		dot = "\x1b[1;91m■\x1b[0m" // red ■
	}

	var path string
	if share := s.sm.CurrentShare(); share == "" {
		path = fmt.Sprintf("\\\\%s\\", s.sm.Host)
	} else {
		cwd := s.sm.Cwd()
		if cwd != "" {
			cwd += "\\"
		}
		path = fmt.Sprintf("\\\\%s\\%s\\%s", s.sm.Host, share, cwd)
	}

	return fmt.Sprintf("%s[\x1b[1;94m%s\x1b[0m]> ", dot, path)
}

// tokenize splits a command line into tokens, honoring double quotes so that
// paths containing spaces can be passed as a single argument.
func tokenize(line string) []string {
	line = strings.TrimRight(line, "\r\n")

	tokens := []string{}
	var current strings.Builder
	inQuotes := false
	hasToken := false

	flush := func() {
		if hasToken {
			tokens = append(tokens, current.String())
			current.Reset()
			hasToken = false
		}
	}

	for _, r := range line {
		switch {
		case r == '"':
			inQuotes = !inQuotes
			hasToken = true
		case (r == ' ' || r == '\t') && !inQuotes:
			flush()
		default:
			current.WriteRune(r)
			hasToken = true
		}
	}
	flush()

	return tokens
}
