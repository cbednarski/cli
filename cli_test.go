// When you write unit tests for your own application, you usually test them in
// the same package. However, when writing a library we need to test the public
// API (which does not have access to private members and follows different
// import rules) to make sure the library works the way an end-user will use it.
package cli_test

import (
	"os"
	"reflect"
	"testing"

	"git.stormbase.io/cbednarski/cli"
)

func TestCLI_Run(t *testing.T) {
	backup := os.Args
	os.Args = []string{"testapp", "testcommand", "testarg1", "testarg2", "testarg3"}
	defer func() { os.Args = backup }()

	var output []string
	expectedOutput := []string{"testarg3", "testarg2", "testarg1"}

	app := cli.CLI{
		Commands: map[string]*cli.Command{
			"testcommand": {
				Run: func(args []string) error {
					// reverse the list of args
					for i := len(args) - 1; i >= 0; i-- {
						output = append(output, args[i])
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(output, expectedOutput) {
		t.Errorf("Expected %#v, found %#v", expectedOutput, output)
	}
}

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
				Help: "We don't support cookies directly, but here's how you can make some:",
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

	t.Run("help with command topic", func(tt *testing.T){
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

	t.Run("help with help-only topic", func (tt *testing.T) {
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
