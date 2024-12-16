package cmd

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"sort"
	"text/template"
)

type Command struct {
	// The command name, used to match the command line argument e.g. clitool <Name>.
	Name string
	// Short is used in the list of commands message e.g. "command does xyz"
	Short string
	// Long is used in the usage message from the help command e.g.:
	// "Usage: <Name> [args]
	//
	// "<Long>"
	Long string
	// Cmd is the function that is called when the command is matched.
	Cmd func()
}

type Cli struct {
	name     string
	long     string
	commands []Command
}

func New(name, long string) *Cli {
	return &Cli{
		name: name,
		long: long,
	}
}

func (c *Cli) AddCmd(cmd Command) {
	c.commands = append(c.commands, cmd)
}

func (c *Cli) Find(cmd string) (Command, bool) {
	for _, v := range c.commands {
		if v.Name == cmd {
			return v, true
		}
	}
	return Command{}, false
}

func (c *Cli) Run() {
	if len(os.Args) < 2 {
		c.usage()
	}
	requestedCmd := os.Args[1]

	// special case for help
	if requestedCmd == "help" {
		c.cmdHelp()
		os.Exit(0)
	}

	cm, ok := c.Find(requestedCmd)
	if !ok {
		c.usage()
	}
	cm.Cmd()
	os.Exit(0)
}

func (c *Cli) cmdHelp() {
	if len(os.Args) < 3 {
		c.usage()
	}
	cmd, ok := c.Find(os.Args[2])
	if ok {
		fmt.Print(cmd.Long)
		os.Exit(0)
	}
	c.usage()
}

// usage returns the usage message, always sorted by command name
// always exits the program with status code 0
func (c *Cli) usage() {
	msg := `{{ .long }}
	
Usage:

	{{ .name }} <command> [arguments]

The commands are:

{{ .commands }}
Use "{{ .name }} help <command>" for more information about a command.

`

	// we need the max length of the command names to align the help message
	longestCmd := slices.MaxFunc(c.commands, func(i, j Command) int {
		return cmp.Compare(len(i.Name), len(j.Name))
	})

	maxCmdLen := len(longestCmd.Name)

	sort.Slice(c.commands, func(i, j int) bool {
		return c.commands[i].Name < c.commands[j].Name
	})

	cmdList := ""
	for _, c := range c.commands {
		cmdList += fmt.Sprintf("\t%-*s%s\n", maxCmdLen+8, c.Name, c.Short)
	}

	m := map[string]any{"long": c.long, "name": c.name, "commands": cmdList}
	template.Must(template.New("usage").Parse(msg)).Execute(os.Stdout, m)
	os.Exit(0)
}
