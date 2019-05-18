package cli

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

var ErrNotImplemented = errors.New("not implemented")

type CLI struct {
	// Name of the program, for help text and command-line examples
	Name string

	// Version displayed for -v and --version
	Version string

	// Header may include a description of the program to be displayed before
	// the command list help output
	Header string

	// Footer may include license and copyright information, a bug report link,
	// project homepage, or other contact information, to be displayed after the
	// command list help output
	Footer string

	//
	Commands map[string]*Command

	//
	HelpTopics map[string]string
}

type Command struct {
	Run func(args []string) error

	// Summary is a terse description of the command shown in the command list.
	// For long-form help text see the Help command.
	Summary string
}

func ListCommandNames(commands map[string]*Command) []string {
	ordered := make([]string, len(commands))
	idx := 0
	for name := range commands {
		ordered[idx] = name
		idx++
	}
	sort.Strings(ordered)
	return ordered
}

func CommandHelp(c *CLI) {
	names := ListCommandNames(c.Commands)

	width := 0
	for _, name := range names {
		if len(name) > width {
			width = len(name)
		}
	}

	fmt.Print(c.Header, "\n\n")
	fmt.Printf("usage: %s [--version] [--help] <command> [<args>]", c.Name)
	fmt.Print("\n\n", "Commands", "\n\n")

	for _, name := range names {
		fmt.Printf("  %s %s   %s\n", c.Name, PadRight(name, width), c.Commands[name].Summary)
	}
	if len(c.HelpTopics) > -1 {
		fmt.Printf("  %s %s   %s\n", c.Name, PadRight("help <topic>", width), "Topic-based help")
	}

	fmt.Print(c.Footer)
}

func Version(c *CLI) {
	if c.Version == "" {
		fmt.Printf("%s version undefined\n", c.Name)
	} else {
		fmt.Printf("%s version %s\n", c.Name, c.Version)
	}
}

func (c *CLI) Run() error {
	commandName, args := ParseArgs(os.Args)

	switch commandName {
	case "":
		CommandHelp(c)
		return nil
	case "--help":
		CommandHelp(c)
		return nil
	case "--version":
		Version(c)
		return nil
	case "help":
		return Help(c.HelpTopics, args)
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

func Help(topics map[string]string, args []string) error {
	// Show help topics if nothing is specified
	if len(args) == 0 {
		fmt.Print("Help Topics\n\n")
		sortedTopics := []string{}
		for topic := range topics {
			sortedTopics = append(sortedTopics, topic)
		}
		sort.Strings(sortedTopics)
		for _, topic := range sortedTopics {
			fmt.Println("  ", topic)
		}
		return nil
	}

	return ErrNotImplemented
}

// ParseArgs separates the command string from any subsequent arguments and
// returns both. It handles cases where command or arguments are not specified.
func ParseArgs(input []string) (command string, args []string) {
	// We're getting os.Args verbatim, so input[0] is always the name of the
	// program. input[1] is the command, and everything else are arguments. If
	// len(input) is less than 2 the user did not type a command so we'll just
	// return here.
	if len(input) > 1 {
		command = input[1]
	}
	if len(input) > 2 {
		args = input[2:]
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
