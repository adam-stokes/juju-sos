// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package cmd_test

import (
	"bytes"
	"fmt"
	"strings"

	gitjujutesting "github.com/juju/testing"
	"launchpad.net/gnuflag"
	gc "launchpad.net/gocheck"

	"github.com/juju/cmd"
	"github.com/juju/cmd/cmdtesting"
)

func initDefenestrate(args []string) (*cmd.SuperCommand, *TestCommand, error) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{Name: "jujutest"})
	tc := &TestCommand{Name: "defenestrate"}
	jc.Register(tc)
	return jc, tc, cmdtesting.InitCommand(jc, args)
}

type SuperCommandSuite struct {
	gitjujutesting.IsolationSuite
}

var _ = gc.Suite(&SuperCommandSuite{})

const helpText = "\n    help\\s+- show help on a command or other topic"
const helpCommandsText = "commands:" + helpText

func (s *SuperCommandSuite) TestDispatch(c *gc.C) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{Name: "jujutest"})
	info := jc.Info()
	c.Assert(info.Name, gc.Equals, "jujutest")
	c.Assert(info.Args, gc.Equals, "<command> ...")
	c.Assert(info.Doc, gc.Matches, helpCommandsText)

	jc, _, err := initDefenestrate([]string{"discombobulate"})
	c.Assert(err, gc.ErrorMatches, "unrecognized command: jujutest discombobulate")
	info = jc.Info()
	c.Assert(info.Name, gc.Equals, "jujutest")
	c.Assert(info.Args, gc.Equals, "<command> ...")
	c.Assert(info.Doc, gc.Matches, "commands:\n    defenestrate - defenestrate the juju"+helpText)

	jc, tc, err := initDefenestrate([]string{"defenestrate"})
	c.Assert(err, gc.IsNil)
	c.Assert(tc.Option, gc.Equals, "")
	info = jc.Info()
	c.Assert(info.Name, gc.Equals, "jujutest defenestrate")
	c.Assert(info.Args, gc.Equals, "<something>")
	c.Assert(info.Doc, gc.Equals, "defenestrate-doc")

	_, tc, err = initDefenestrate([]string{"defenestrate", "--option", "firmly"})
	c.Assert(err, gc.IsNil)
	c.Assert(tc.Option, gc.Equals, "firmly")

	_, tc, err = initDefenestrate([]string{"defenestrate", "gibberish"})
	c.Assert(err, gc.ErrorMatches, `unrecognized args: \["gibberish"\]`)

	// --description must be used on it's own.
	_, _, err = initDefenestrate([]string{"--description", "defenestrate"})
	c.Assert(err, gc.ErrorMatches, `unrecognized args: \["defenestrate"\]`)
}

func (s *SuperCommandSuite) TestRegister(c *gc.C) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{Name: "jujutest"})
	jc.Register(&TestCommand{Name: "flip"})
	jc.Register(&TestCommand{Name: "flap"})
	badCall := func() { jc.Register(&TestCommand{Name: "flap"}) }
	c.Assert(badCall, gc.PanicMatches, "command already registered: flap")
}

func (s *SuperCommandSuite) TestRegisterAlias(c *gc.C) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{Name: "jujutest"})
	jc.Register(&TestCommand{Name: "flip", Aliases: []string{"flap", "flop"}})

	info := jc.Info()
	c.Assert(info.Doc, gc.Equals, `commands:
    flap - alias for flip
    flip - flip the juju
    flop - alias for flip
    help - show help on a command or other topic`)
}

func (s *SuperCommandSuite) TestInfo(c *gc.C) {
	commandsDoc := `commands:
    flapbabble - flapbabble the juju
    flip       - flip the juju`

	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{
		Name:    "jujutest",
		Purpose: "to be purposeful",
		Doc:     "doc\nblah\ndoc",
	})
	info := jc.Info()
	c.Assert(info.Name, gc.Equals, "jujutest")
	c.Assert(info.Purpose, gc.Equals, "to be purposeful")
	// info doc starts with the jc.Doc and ends with the help command
	c.Assert(info.Doc, gc.Matches, jc.Doc+"(.|\n)*")
	c.Assert(info.Doc, gc.Matches, "(.|\n)*"+helpCommandsText)

	jc.Register(&TestCommand{Name: "flip"})
	jc.Register(&TestCommand{Name: "flapbabble"})
	info = jc.Info()
	c.Assert(info.Doc, gc.Matches, jc.Doc+"\n\n"+commandsDoc+helpText)

	jc.Doc = ""
	info = jc.Info()
	c.Assert(info.Doc, gc.Matches, commandsDoc+helpText)
}

type testVersionFlagCommand struct {
	cmd.CommandBase
	version string
}

func (c *testVersionFlagCommand) Info() *cmd.Info {
	return &cmd.Info{Name: "test"}
}

func (c *testVersionFlagCommand) SetFlags(f *gnuflag.FlagSet) {
	f.StringVar(&c.version, "version", "", "")
}

func (c *testVersionFlagCommand) Run(_ *cmd.Context) error {
	return nil
}

func (s *SuperCommandSuite) TestVersionFlag(c *gc.C) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{
		Name:    "jujutest",
		Purpose: "to be purposeful",
		Doc:     "doc\nblah\ndoc",
		Version: "111.222.333",
	})
	testVersionFlagCommand := &testVersionFlagCommand{}
	jc.Register(testVersionFlagCommand)

	var stdout, stderr bytes.Buffer
	ctx := &cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	// baseline: juju version
	code := cmd.Main(jc, ctx, []string{"version"})
	c.Check(code, gc.Equals, 0)
	baselineStderr := stderr.String()
	baselineStdout := stdout.String()
	stderr.Reset()
	stdout.Reset()

	// juju --version output should match that of juju version.
	code = cmd.Main(jc, ctx, []string{"--version"})
	c.Check(code, gc.Equals, 0)
	c.Assert(stderr.String(), gc.Equals, baselineStderr)
	c.Assert(stdout.String(), gc.Equals, baselineStdout)
	stderr.Reset()
	stdout.Reset()

	// juju test --version should update testVersionFlagCommand.version,
	// and there should be no output. The --version flag on the 'test'
	// subcommand has a different type to the "juju --version" flag.
	code = cmd.Main(jc, ctx, []string{"test", "--version=abc.123"})
	c.Check(code, gc.Equals, 0)
	c.Assert(stderr.String(), gc.Equals, "")
	c.Assert(stdout.String(), gc.Equals, "")
	c.Assert(testVersionFlagCommand.version, gc.Equals, "abc.123")
}

func (s *SuperCommandSuite) TestVersionNotProvided(c *gc.C) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{
		Name:    "jujutest",
		Purpose: "to be purposeful",
		Doc:     "doc\nblah\ndoc",
	})
	var stdout, stderr bytes.Buffer
	ctx := &cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	// juju version
	baselineCode := cmd.Main(jc, ctx, []string{"version"})
	c.Check(baselineCode, gc.Not(gc.Equals), 0)
	c.Assert(stderr.String(), gc.Equals, "error: unrecognized command: jujutest version\n")
	stderr.Reset()
	stdout.Reset()

	// juju --version
	code := cmd.Main(jc, ctx, []string{"--version"})
	c.Check(code, gc.Equals, baselineCode)
	c.Assert(stderr.String(), gc.Equals, "error: flag provided but not defined: --version\n")
}

func (s *SuperCommandSuite) TestLogging(c *gc.C) {
	sc := cmd.NewSuperCommand(cmd.SuperCommandParams{
		UsagePrefix: "juju",
		Name:        "command",
		Log:         &cmd.Log{},
	})
	sc.Register(&TestCommand{Name: "blah"})
	ctx := cmdtesting.Context(c)
	code := cmd.Main(sc, ctx, []string{"blah", "--option", "error", "--debug"})
	c.Assert(code, gc.Equals, 1)
	c.Assert(bufferString(ctx.Stderr), gc.Matches, `^.* ERROR .* BAM!\n`)
}

func (s *SuperCommandSuite) TestNotifyRun(c *gc.C) {
	notifyTests := []struct {
		usagePrefix string
		name string
		expectName string
	}{
		{"juju", "juju", "juju"},
		{"something", "else", "something else"},
		{"", "juju", "juju"},
		{"", "myapp", "myapp"},
	}
	for i, test := range notifyTests {
		c.Logf("test %d. %q %q", i, test.usagePrefix, test.name)
		notifyName := ""
		sc := cmd.NewSuperCommand(cmd.SuperCommandParams{
			UsagePrefix: test.usagePrefix,
			Name:        test.name,
			NotifyRun: func(name string) {
				notifyName = name
			},
		})
		sc.Register(&TestCommand{Name: "blah"})
		ctx := cmdtesting.Context(c)
		code := cmd.Main(sc, ctx, []string{"blah", "--option", "error"})
		c.Assert(bufferString(ctx.Stderr), gc.Matches, "")
		c.Assert(code, gc.Equals, 1)
		c.Assert(notifyName, gc.Equals, test.expectName)
	}
}

func (s *SuperCommandSuite) TestDescription(c *gc.C) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{Name: "jujutest", Purpose: "blow up the death star"})
	jc.Register(&TestCommand{Name: "blah"})
	ctx := cmdtesting.Context(c)
	code := cmd.Main(jc, ctx, []string{"blah", "--description"})
	c.Assert(code, gc.Equals, 0)
	c.Assert(bufferString(ctx.Stdout), gc.Equals, "blow up the death star\n")
}

func (s *SuperCommandSuite) TestHelp(c *gc.C) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{Name: "jujutest"})
	jc.Register(&TestCommand{Name: "blah"})
	ctx := cmdtesting.Context(c)
	code := cmd.Main(jc, ctx, []string{"blah", "--help"})
	c.Assert(code, gc.Equals, 0)
	stripped := strings.Replace(bufferString(ctx.Stdout), "\n", "", -1)
	c.Assert(stripped, gc.Matches, ".*usage: jujutest blah.*blah-doc.*")
}

func (s *SuperCommandSuite) TestHelpWithPrefix(c *gc.C) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{Name: "jujutest", UsagePrefix: "juju"})
	jc.Register(&TestCommand{Name: "blah"})
	ctx := cmdtesting.Context(c)
	code := cmd.Main(jc, ctx, []string{"blah", "--help"})
	c.Assert(code, gc.Equals, 0)
	stripped := strings.Replace(bufferString(ctx.Stdout), "\n", "", -1)
	c.Assert(stripped, gc.Matches, ".*usage: juju jujutest blah.*blah-doc.*")
}

func NewSuperWithCallback(callback func(*cmd.Context, string, []string) error) cmd.Command {
	return cmd.NewSuperCommand(cmd.SuperCommandParams{
		Name:            "jujutest",
		Log:             &cmd.Log{},
		MissingCallback: callback,
	})
}

func (s *SuperCommandSuite) TestMissingCallback(c *gc.C) {
	var calledName string
	var calledArgs []string

	callback := func(ctx *cmd.Context, subcommand string, args []string) error {
		calledName = subcommand
		calledArgs = args
		return nil
	}

	code := cmd.Main(
		NewSuperWithCallback(callback),
		cmdtesting.Context(c),
		[]string{"foo", "bar", "baz", "--debug"})
	c.Assert(code, gc.Equals, 0)
	c.Assert(calledName, gc.Equals, "foo")
	c.Assert(calledArgs, gc.DeepEquals, []string{"bar", "baz", "--debug"})
}

func (s *SuperCommandSuite) TestMissingCallbackErrors(c *gc.C) {
	callback := func(ctx *cmd.Context, subcommand string, args []string) error {
		return fmt.Errorf("command not found %q", subcommand)
	}

	ctx := cmdtesting.Context(c)
	code := cmd.Main(NewSuperWithCallback(callback), ctx, []string{"foo"})
	c.Assert(code, gc.Equals, 1)
	c.Assert(cmdtesting.Stdout(ctx), gc.Equals, "")
	c.Assert(cmdtesting.Stderr(ctx), gc.Equals, "ERROR command not found \"foo\"\n")
}

func (s *SuperCommandSuite) TestMissingCallbackContextWiredIn(c *gc.C) {
	callback := func(ctx *cmd.Context, subcommand string, args []string) error {
		fmt.Fprintf(ctx.Stdout, "this is std out")
		fmt.Fprintf(ctx.Stderr, "this is std err")
		return nil
	}

	ctx := cmdtesting.Context(c)
	code := cmd.Main(NewSuperWithCallback(callback), ctx, []string{"foo", "bar", "baz", "--debug"})
	c.Assert(code, gc.Equals, 0)
	c.Assert(cmdtesting.Stdout(ctx), gc.Equals, "this is std out")
	c.Assert(cmdtesting.Stderr(ctx), gc.Equals, "this is std err")
}

func (s *SuperCommandSuite) TestSupercommandAliases(c *gc.C) {
	jc := cmd.NewSuperCommand(cmd.SuperCommandParams{
		Name:        "jujutest",
		UsagePrefix: "juju",
	})
	sub := cmd.NewSuperCommand(cmd.SuperCommandParams{
		Name:        "jubar",
		UsagePrefix: "juju jujutest",
		Aliases:     []string{"jubaz", "jubing"},
	})
	info := sub.Info()
	c.Check(info.Aliases, gc.DeepEquals, []string{"jubaz", "jubing"})
	jc.Register(sub)
	for _, name := range []string{"jubar", "jubaz", "jubing"} {
		c.Logf("testing command name %q", name)
		ctx := cmdtesting.Context(c)
		code := cmd.Main(jc, ctx, []string{name, "--help"})
		c.Assert(code, gc.Equals, 0)
		stripped := strings.Replace(bufferString(ctx.Stdout), "\n", "", -1)
		c.Assert(stripped, gc.Matches, ".*usage: juju jujutest jubar.*aliases: jubaz, jubing")
	}
}
