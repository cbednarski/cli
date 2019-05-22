// When you write unit tests for your own application, you usually test them in
// the same package. However, when writing a library we need to test the public
// API (which does not have access to private members and follows different
// import rules) to make sure the library works the way an end-user will use it.
package cli_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"git.stormbase.io/cbednarski/cli"
)

func TestSortedCommandNames(t *testing.T) {
	commands := map[string]*cli.Command{
		"map":    {},
		"filter": {},
		"reduce": {},
		"find":   {},
		"keys":   {},
	}

	expected := []string{
		"filter",
		"find",
		"keys",
		"map",
		"reduce",
	}

	sorted := cli.SortedCommandNames(commands)

	if !reflect.DeepEqual(expected, sorted) {
		t.Errorf("Expected %#v found %#v", expected, sorted)
	}
}

func TestCommandHelp(t *testing.T) {
	commands := map[string]*cli.Command{
		"mix": {
			Summary: "incorporate your ingredients",
		},
		"bake": {
			Summary: "heat things up",
		},
		"eat": {
			Summary: "enjoy delicious cake!",
		},
		"cleanup": {
			Hidden: true,
		},
	}

	app := &cli.CLI{
		Header:   "It's time to enjoy something tasty",
		Name:     "cake",
		Footer:   "Did you like it? Make another and share it with your friends!",
		Commands: commands,
	}

	expectedOutput := `It's time to enjoy something tasty

usage: cake [--version] [--help] <command> [<args>]

Commands

  cake bake   heat things up
  cake eat    enjoy delicious cake!
  cake mix    incorporate your ingredients
  cake help   List help topics

Did you like it? Make another and share it with your friends!
`

	output := cli.CommandHelp(app)

	if output != expectedOutput {
		t.Errorf("--- Expected Output ---\n%s\n--- Actual Output ---\n%s\n", expectedOutput, output)
	}

	// Verify we still get the correct formatting when header and footer are
	// defined with extra newlines at beginning / end
	app.Header = `
It's time to enjoy something tasty
`
	app.Footer = `
Did you like it? Make another and share it with your friends!
`

	output = cli.CommandHelp(app)

	if output != expectedOutput {
		t.Errorf("--- Expected Output ---\n%s\n--- Actual Output ---\n%s\n", expectedOutput, output)
	}

}

func TestVersion(t *testing.T) {
	app := &cli.CLI{
		Name: "chocolate",
	}

	expectedUndefined := "chocolate version undefined"
	actualUndefined := cli.Version(app)

	if expectedUndefined != actualUndefined {
		t.Errorf("Expected %q, found %q", expectedUndefined, actualUndefined)
	}

	app.Version = "0.1.0"

	expectedOhOneOh := "chocolate version 0.1.0"
	actualOnOneOh := cli.Version(app)

	if expectedOhOneOh != actualOnOneOh {
		t.Errorf("Expected %q, found %q", expectedOhOneOh, actualOnOneOh)
	}
}

func TestHelp(t *testing.T) {
	app := &cli.CLI{
		Name: "testapp",
		Commands: map[string]*cli.Command{
			"candy": {
				Help: "There are many tasty varieties of candy. Here are some of the flavors you can choose from.",
			},
			"cookies": {
				Help:     "We don't support cookies directly, but here's how you can make some:",
				HelpOnly: true,
			},
		},
	}

	t.Run("bare help command", func(tt *testing.T) {
		output, err := cli.Help(app, []string{})
		if err != nil {
			tt.Fatal(err)
		}

		expectedOutput := `usage: testapp help <topic>

Help Topics

  candy (command)
  cookies
`

		if output != expectedOutput {
			tt.Errorf("--- Expected Output ---\n%s\n--- Actual Output ---\n%s\n", expectedOutput, output)
		}
	})

	t.Run("help with command topic", func(tt *testing.T) {
		output, err := cli.Help(app, []string{"candy"})
		if err != nil {
			tt.Fatal(err)
		}

		expectedOutput := `candy Command Help

There are many tasty varieties of candy. Here are some of the flavors you can choose from.
`

		if output != expectedOutput {
			tt.Errorf("--- Expected Output ---\n%s\n--- Actual Output ---\n%s\n", expectedOutput, output)
		}
	})

	t.Run("help with help-only topic", func(tt *testing.T) {
		output, err := cli.Help(app, []string{"cookies"})
		if err != nil {
			tt.Fatal(err)
		}

		expectedOutput := `cookies Help

We don't support cookies directly, but here's how you can make some:
`

		if output != expectedOutput {
			tt.Errorf("--- Expected Output ---\n%s\n--- Actual Output ---\n%s\n", expectedOutput, output)
		}

	})

	t.Run("missing help topic", func(tt *testing.T) {
		_, err := cli.Help(app, []string{"cake"})
		expectedError := "unknown help topic 'cake'"
		if err.Error() != expectedError {
			t.Errorf("Expected %q, found %q", expectedError, err.Error())
		}
	})
}

func TestParseArgs(t *testing.T) {
	type TestCase struct {
		Input           []string
		ExpectedCommand string
		ExpectedArgs    []string
	}

	cases := []TestCase{
		{
			Input:           []string{"cat", "file1", "file2"},
			ExpectedCommand: "cat",
			ExpectedArgs:    []string{"file1", "file2"},
		},
		{
			Input:           []string{},
			ExpectedCommand: "",
			ExpectedArgs:    []string{},
		},
		{
			Input:           []string{"cat"},
			ExpectedCommand: "cat",
			ExpectedArgs:    []string{},
		},
		{
			Input:           []string{"cat", "file1"},
			ExpectedCommand: "cat",
			ExpectedArgs:    []string{"file1"},
		},
	}

	for _, testCase := range cases {
		command, args := cli.ParseArgs(testCase.Input)

		if command != testCase.ExpectedCommand {
			t.Errorf("Expected %q, found %q", testCase.ExpectedCommand, command)
		}

		if !reflect.DeepEqual(args, testCase.ExpectedArgs) {
			t.Errorf("Expected %#v, found %#v", testCase.ExpectedArgs, args)
		}
	}
}

func TestPadRight(t *testing.T) {
	type TestCase struct {
		Str      string
		Width    int
		Expected string
	}

	cases := []TestCase{
		{
			Str:      "candy",
			Width:    10,
			Expected: "candy     ",
		},
		{
			Str:      "",
			Width:    -5,
			Expected: "",
		},
		{
			Str:      "waka",
			Width:    2,
			Expected: "waka",
		},
	}

	for _, testCase := range cases {
		actual := cli.PadRight(testCase.Str, testCase.Width)
		if actual != testCase.Expected {
			t.Errorf("Expected %q, found %q with input (%q, %d)", testCase.Expected, actual, testCase.Str, testCase.Width)
		}
	}
}

func redirectIO() (cleanup func(), stdout *os.File) {
	ogArgs := os.Args
	ogStdout := os.Stdout

	cleanup = func() {
		stdout.Close()

		os.Args = ogArgs
		os.Stdout = ogStdout
	}

	var err error

	stdout, err = ioutil.TempFile("", "cli-test-stdout")
	if err != nil {
		panic(err)
	}
	os.Stdout = stdout

	return
}

//func TestExitWithError(t *testing.T) {
//	cleanup, _, stderr := redirectIO()
//
//	err := fmt.Errorf("pie pie pie!")
//	cli.ExitWithError(err)
//
//	cleanup()
//
//	data, err := ioutil.ReadFile(stderr.Name())
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if string(data) != err.Error() {
//		t.Errorf("Expected %q, found %q", err.Error(), string(data))
//	}
//}

func TestCLI_Run(t *testing.T) {
	app := &cli.CLI{
		Commands: map[string]*cli.Command{
			"reverse": {
				Summary: "reverse the arguments",
				Run: func(args []string) error {
					// reverse the list of args
					var output []string
					for i := len(args) - 1; i >= 0; i-- {
						output = append(output, args[i])
					}
					fmt.Println(strings.Join(output, " "))
					return nil
				},
				Help: "All arguments passed to the command will be displayed in reverse order",
			},
			"todo": {},
			"error": {
				Run: func(args []string) error {
					return fmt.Errorf("error error error!")
				},
			},
		},
	}

	t.Run("app name", func(t *testing.T) {
		cleanup, _ := redirectIO()
		defer cleanup()

		expectedAppName := "testapp"
		os.Args = []string{expectedAppName}
		if err := app.Run(); err != nil {
			t.Fatal(err)
		}

		if app.Name != expectedAppName {
			t.Errorf("Expected %q, found %q", expectedAppName, app.Name)
		}
	})

	t.Run("basic invocation", func(t *testing.T) {
		cleanup, stdout := redirectIO()
		defer cleanup()

		os.Args = []string{"testapp"}
		if err := app.Run(); err != nil {
			t.Fatal(err)
		}

		cleanup() // Cleanup to flush stdout/err to disk
		output, err := ioutil.ReadFile(stdout.Name())
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := cli.CommandHelp(app)

		if string(output) != expectedOutput {
			t.Errorf("--- Expected Output ---\n%s\n--- Actual Output ---\n%s\n", expectedOutput, string(output))
		}
	})

	t.Run("--help", func(t *testing.T) {
		cleanup, stdout := redirectIO()
		defer cleanup()

		os.Args = []string{"testapp", "--help"}
		if err := app.Run(); err != nil {
			t.Fatal(err)
		}

		cleanup() // Cleanup to flush stdout/err to disk
		output, err := ioutil.ReadFile(stdout.Name())
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := cli.CommandHelp(app)

		if string(output) != expectedOutput {
			t.Errorf("--- Expected Output ---\n%s\n--- Actual Output ---\n%s\n", expectedOutput, string(output))
		}
	})

	t.Run("command invocation", func(t *testing.T) {
		cleanup, stdout := redirectIO()
		defer cleanup()

		os.Args = []string{"testapp", "reverse", "testarg1", "testarg2", "testarg3"}
		expectedOutput := "testarg3 testarg2 testarg1\n"

		if err := app.Run(); err != nil {
			t.Fatal(err)
		}

		cleanup() // Cleanup to flush stdout/err to disk
		output, err := ioutil.ReadFile(stdout.Name())
		if err != nil {
			t.Fatal(err)
		}

		if string(output) != expectedOutput {
			t.Errorf("Expected %#v, found %#v", expectedOutput, string(output))
		}
	})

	t.Run("--version", func(t *testing.T) {
		cleanup, stdout := redirectIO()
		defer cleanup()

		os.Args = []string{"testapp", "--version"}
		expectedOutput := "testapp version undefined\n"

		if err := app.Run(); err != nil {
			t.Fatal(err)
		}

		cleanup() // Cleanup to flush stdout/err to disk
		output, err := ioutil.ReadFile(stdout.Name())
		if err != nil {
			t.Fatal(err)
		}

		if string(output) != expectedOutput {
			t.Errorf("Expected %q, found %q", expectedOutput, string(output))
		}
	})

	t.Run("help", func(t *testing.T) {
		cleanup, stdout := redirectIO()
		defer cleanup()

		os.Args = []string{"testapp", "help"}
		expectedOutput, err := cli.Help(app, []string{})
		if err != nil {
			t.Fatal(err)
		}

		if err := app.Run(); err != nil {
			t.Fatal(err)
		}

		cleanup() // Cleanup to flush stdout/err to disk
		output, err := ioutil.ReadFile(stdout.Name())
		if err != nil {
			t.Fatal(err)
		}

		if string(output) != expectedOutput {
			t.Errorf("--- Expected Output ---\n%s\n--- Actual Output ---\n%s\n", expectedOutput, string(output))
		}
	})

	t.Run("invalid command", func(t *testing.T) {
		os.Args = []string{"testapp", "cookies"}

		err := app.Run()
		if err == nil {
			t.Error("expected error")
		}

		expectedOutput := "'cookies' is not a testapp command. See 'testapp --help'."

		if err.Error() != expectedOutput {
			t.Errorf("Expected %q, found %s", expectedOutput, err.Error())
		}
	})

	t.Run("command not implemented", func(t *testing.T) {
		os.Args = []string{"testapp", "todo"}

		err := app.Run()
		if err == nil {
			t.Error("expected error")
		}

		expectedOutput := "not implemented"

		if err.Error() != expectedOutput {
			t.Errorf("Expected %q, found %s", expectedOutput, err.Error())
		}
	})

	t.Run("command error", func(t *testing.T) {
		os.Args = []string{"testapp", "error"}

		err := app.Run()
		if err == nil {
			t.Error("expected error")
		}

		expectedOutput := "error error error!"

		if err.Error() != expectedOutput {
			t.Errorf("Expected %q, found %s", expectedOutput, err.Error())
		}
	})

	t.Run("invalid program name", func(t *testing.T) {
		app.Name = "has a space"

		err := app.Run()
		if err == nil {
			t.Error("expected error")
		}

		expectedOutput := `program name ("has a space") must not contain spaces, try renaming the binary`

		if err.Error() != expectedOutput {
			t.Errorf("Expected %q, found %q", expectedOutput, err.Error())
		}
	})

}
