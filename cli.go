// Package cli provides a way to quickly build command-line interfaces like
// those seen in git, docker, go, vagrant, and other contemporary tools without
// any boilerplate. For example:
//
//	program command arg1 arg2 arg3
//
// cli's types are designed to have safe defaults so you can leave them empty
// and quickly prototype a program, filling in additional details as you go.
// Here is a simple example program you can use as a starting point:
//
//	func main() {
//		commands := map[string]*cli.Command{
//			"go": {},
//		}
//
//		app := &cli.CLI{
//			Commands: commands,
//		}
//
//		if err := app.Run(); err != nil {
//			cli.ExitWithError(err)
//		}
//	}
//
// When a program is run without arguments, cli will display command help and
// exit. As a result, cli is not particularly well-suited to building
// traditional unix tools like ls, top, or grep. To build interfaces like these
// you may prefer Go's stdlib flag package.
package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	ErrNotImplemented   = errors.New("not implemented")
	ErrTooManyArguments = errors.New("too many arguments")
)

// CLI is used to define parts of a command-line application, including the list
// of Commands that a user may call. Every other field is optional, but if
// defined it will be used to populate output of the built-in --help, --version
// and help commands.
type CLI struct {
	// Name of the program, for help text and command-line examples. This must
	// be a valid filename on every target system and MUST NOT CONTAIN SPACES.
	// If you do not set this it will be set automatically using this snippet:
	//
	//	filepath.Base(os.Args[0])
	Name string

	// Version displayed when the program is invoked with --version.
	Version string

	// Header defines arbitrary text that is displayed above the command list
	// when the user invokes --help or runs the program without arguments.
	//
	// Header is a great place to direct users to an introductory help topic,
	// an installation or configuration guide, or other setup instructions.
	Header string

	// Footer defines arbitrary text that will be displayed below the command
	// list when the user invokes --help or runs the program without arguments.
	// It is a good place to include license and copyright information, bug
	// report or support links, project homepage, etc.
	Footer string

	// Commands are invoked by their map key.
	Commands map[string]*Command
}

// Run starts by parsing os.Args[1:] and uses the first "argument" to the
// program as the command that will be invoked.
//
// CLI only parses --version when the program is invoked with no commands so you
// are free to use a --version flag in your own UI and it will not collide.
//
//TODO
// CLI will parse --help under any command and will display the command list,
// subcommand list, or command help, depending on context.
//
// The 'help' command is only parsed after the program name and will not be
// invoked when calling commands or subcommands, so you may use this as an
// argument or other input to your program.
//
// All Commands should be specified before Run is called. Modifying CLI or
// Commands after calling Run will produce undefined behavior.
func (c *CLI) Run() error {
	commandName, args := ParseArgs(os.Args[1:])

	// Set a default name for the program in case the user forgot to set one.
	// This also automatically detects the program name if the binary is renamed
	// so it's a decent default behavior.
	if c.Name == "" {
		c.Name = filepath.Base(os.Args[0])
	}

	// Enforce no spaces in command and program names because this will break
	// all kinds of stuff. There are technically other ways to break the program
	// (non-printing characters, for example) but to be as permissive as
	// possible for UTF-8 names we won't validate anything else.
	if strings.ContainsAny(c.Name, " \n\t") {
		// This could happen because of user behavior so we'll error instead of
		// panicking and give the user a chance to fix it.
		return fmt.Errorf("program name (%q) must not contain spaces, try renaming the binary", c.Name)
	}
	for name := range c.Commands {
		if strings.ContainsAny(name, " \n\t") {
			// This is a programmer error and there's no way for the user to fix
			// it so we'll just panic.
			panic(fmt.Sprintf("command names (%q) must not contain spaces", name))
		}
	}

	switch commandName {
	case "":
		fmt.Print(CommandHelp(c))
		return nil
	case "--help":
		fmt.Print(CommandHelp(c))
		return nil
	case "--version":
		fmt.Println(Version(c))
		return nil
	case "help":
		output, err := Help(c, args)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	}

	command, ok := c.Commands[commandName]
	if !ok {
		return fmt.Errorf("'%s' is not a %s command. See '%s --help'.", commandName, c.Name, c.Name)
	}

	if command.Run == nil {
		return ErrNotImplemented
	}

	if err := command.Run(args); err != nil {
		return err
	}

	return nil
}

// Command defines a CLI command that may be invoked by the key name in
// CLI.Commands. Command names MUST NOT CONTAIN SPACES. A space in a command
// name will result in a panic.
//
// It should always be safe to pass an empty Command struct to any of the CLI
// functions. This is not interesting, but it should never result in a crash.
type Command struct {
	// Run is passed arguments by cli.Run(). Any error returned by Run will be
	// shown to the user
	Run func(args []string) error

	// Summary is a terse description of the command shown in the command list.
	// For long-form help text see the Help command.
	//
	// If Summary is set to the cli.Hidden constant then the command will not be
	// displayed in the command list or help output but will still function as a
	// normal command.
	//
	// If Summary is set to the cli.HelpOnly constant then the command will be
	// treated as a help topic only, accessible by the help command but not as
	// a normal command.
	Summary string

	// Help bears a long-form help page. It may be associated with a command or
	// displayed stand-alone, and will be displayed using the help command.
	//
	// When the help command is invoked with no arguments, it will produce a
	// list including each Command item with a non-empty Help string.
	Help string

	// Hidden commands may still be invoked as normal, but will be excluded from
	// the command list. This is useful for deprecating commands or creating
	// additional or special commands that are not part of the UI.
	//
	// A hidden command that has Help specified will still be displayed in the
	// list of help topics, but Summary will never be used.
	Hidden bool

	// HelpOnly commands are used to display additional information via the help
	// command, but cannot actually be invoked. These are useful for displaying
	// additional help topics to the user, such as installation or configuration
	// instructions.
	HelpOnly bool

	//TODO
	// Commands is used to implement subcommands invoked by calling the program
	// name followed by the command, and subsequently the subcommand. These may
	// be nested to any arbitrary depth.
	//
	// Important Caveats
	//
	// Any command that has subcommands cannot be invoked directly. Instead,
	// command help will be displayed that lists the available subcommands. This
	// prevents collisions between arguments and subcommand names.
	//
	// While subcommands are analyzed recursively, the tree is analyzed only
	// once when the CLI arguments are initially parsed and as a result the
	// program cannot dynamically add subcommands on-the-fly.
	//Commands map[string]*Command
}

// SortedCommandNames returns a list of command names in lexical order.
func SortedCommandNames(commands map[string]*Command) []string {
	ordered := make([]string, len(commands))
	idx := 0
	for name := range commands {
		ordered[idx] = name
		idx++
	}
	sort.Strings(ordered)
	return ordered
}

// CommandHelp
func CommandHelp(c *CLI) (output string) {
	names := SortedCommandNames(c.Commands)

	width := 0
	for _, name := range names {
		// Skip hidden and help-only commands
		if !c.Commands[name].Hidden && !c.Commands[name].HelpOnly && len(name) > width {
			width = len(name)
		}
	}

	header := c.Header

	if header != "" {
		// Checking for a newline character at the start and end allows us to
		// achieve the desired output format with a variety of natural string
		// definitions in Go source code. For example, both of the following
		// will produce the correct output format:
		//
		//	Header = "text goes here"
		//
		//	Header = `
		//	text goes here
		//	`
		//
		// We use the same approach for Footer, below.
		if strings.HasPrefix(header, "\n") {
			header = header[1:]
		}
		output += header
		if strings.HasSuffix(header, "\n") {
			output += "\n"
		} else {
			output += "\n\n"
		}
	}

	output += fmt.Sprintf("usage: %s [--version] [--help] <command> [<args>]", c.Name)
	output += fmt.Sprint("\n\n", "Commands", "\n\n")

	for _, name := range names {
		// Skip hidden and help-only commands
		if !c.Commands[name].Hidden && !c.Commands[name].HelpOnly {
			output += fmt.Sprintf("  %s %s   %s\n", c.Name, PadRight(name, width), c.Commands[name].Summary)
		}
	}
	if len(c.Commands) > -1 {
		output += fmt.Sprintf("  %s %s   %s\n", c.Name, PadRight("help", width), "List help topics")
	}

	if c.Footer != "" {
		if !strings.HasPrefix(c.Footer, "\n") {
			output += "\n"
		}
		output += c.Footer
		if !strings.HasSuffix(c.Footer, "\n") {
			output += "\n"
		}
	}

	return
}

func Version(c *CLI) string {
	if c.Version == "" {
		return fmt.Sprintf("%s version undefined", c.Name)
	}
	return fmt.Sprintf("%s version %s", c.Name, c.Version)
}

func Help(c *CLI, args []string) (output string, err error) {
	switch len(args) {
	case 0:
		// Show help topics if nothing is specified
		output += fmt.Sprintf("usage: %s help <topic>\n\nHelp Topics\n\n", c.Name)
		names := SortedCommandNames(c.Commands)
		for _, topic := range names {
			if !c.Commands[topic].Hidden && c.Commands[topic].Help != "" {
				if c.Commands[topic].HelpOnly {
					output += fmt.Sprintf("  %s\n", topic)
				} else {
					output += fmt.Sprintf("  %s (command)\n", topic)
				}
			}
		}
	case 1:
		// Show help for a single topic
		topic := args[0]
		command, ok := c.Commands[topic]
		if !ok {
			err = fmt.Errorf("unknown help topic '%s'", topic)
			return
		}

		// Show the help topic
		output += topic
		// Show "Command Help" if the help topic is attached to a normal command
		if !command.HelpOnly {
			output += " Command"
		}
		output += " Help\n\n"
		output += command.Help

		// Ensure newline at end of output
		if !strings.HasSuffix(command.Help, "\n") {
			output += fmt.Sprint("\n")
		}
	default:
		// TODO tweak this for subcommand help
		err = ErrTooManyArguments
	}

	return
}

// ParseArgs separates the command string from any subsequent arguments and
// returns both. It handles cases where command or arguments are not specified.
func ParseArgs(input []string) (command string, args []string) {
	args = []string{}
	if len(input) > 0 {
		command = input[0]
	}
	if len(input) > 1 {
		args = input[1:]
	}

	return
}

// ExitWithError writes the error to stderr and halts with exit code 1. It is
// used in main() to handle errors returned from Run() or WrappedMain(), such as
//
//	func main() {
//		...
//
//		if err := cli.Run(commands, helpTopics); err != nil {
//			ExitWithError(err)
//		}
//	}
func ExitWithError(err error) {
	_, _ = os.Stderr.WriteString("error: ")
	_, _ = os.Stderr.WriteString(err.Error())
	_, _ = os.Stderr.WriteString("\n")
	os.Exit(1)
}

// PadRight will append spaces to a string until it reaches the specified width
func PadRight(str string, width int) string {
	if len(str) >= width {
		return str
	}

	return str + strings.Repeat(" ", width-len(str))
}
