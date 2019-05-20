# CLI

CLI is designed to help you quickly build _verb-noun_ interfaces like those found in `git`, `docker`, `go`, `lxc`, `packer`, and other contemporary programs.

    program-name command-name arg1 arg2 arg3 ...

CLI is designed to have valid defaults so you may skip defining fields you don't need or care about. For more details, see the package documentation.

## Built-in Features

CLI includes a few built-in features that every program should have

- `program-name --version` displays the program's version
- `program-name --help` displays **command help** (a list of available commands)
- `program-name help` can be invoked to show help about commands or other topics 
- Running `program-name` without additional parameters will display command help

## Examples

### Defining Commands

Let's say we were building a git clone. We could start out like so:

	func main() {
		commands := map[string]*cli.Command{
			"clone": {
				Summary: "Clone a repository into a new directory",
				Run: func(args []string) error {
					if len(args) != 1 {
						return errors.New("")
					}
					remote := args[0]
					// TODO implement some cloning
					return nil
				},
				Help: "Verbose documentation ..."
			},
			"init": {
				Summary:"Create an empty Git repository or reinitialize an existing one",
				Run: func(args []string) error {
					// create a new repository
					if err := os.MkdirAll(".git"); err != nil {
						return err
					}
					// TODO implement some initialization
					return nil
				},
			},
		}

		app := &cli.CLI{
			Name: "git",
			Commands: commands,
		}

		if err := app.Run(); err != nil {
		    // If there is an error, write to stderr and exit(1)
			cli.ExitWithError(err)
		}
	}

All Commands should be specified before calling `CLI.Run()`.

### Defining the Application

CLI will generate command help based on the list of commands specified in the program. It is also useful to specify the program name and version.

	app := &cli.CLI{
		Name:     "git",
		Version:  "0.0.1",
		Header:   "the stupid content tracker",
		Footer:   "Copyright some year, some people, licensed under GPLv2",
		Commands: map[string]*cli.Command{},
	}

`Name` is used to reference the name of the program in command help and other places. CLI will try to detect the name if you don't specify this field. `Version` indicates the current version of your program. If it is not specified the CLI will report `undefined`.

`Header` and `Footer` are used to show additional text at the top and bottom of the command help output. Typically, `Header` should include a brief description of the program and any essential information for first use, while `Footer` may be used for copyright and license information, project homepage, bug report URLs, etc.

`Commands` holds the map of commands that users can run. `--version`, `--help`, and `help` are reserved.
