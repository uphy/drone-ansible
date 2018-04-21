package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type (
	//Config defined the ansible configuration params
	Build struct {
		Path string
		SHA  string
		Tag  string
	}

	//Config defined the ansible configuration params
	Config struct {
		InventoryPath string
		Inventories   []string
		Playbook      string
		SSHKey        string
		SSHUser       string
		SSHPassphrase string
		Debug         bool
	}

	// Plugin defines the Ansible plugin parameters.
	Plugin struct {
		env *environment
	}
)

func New(build *Build, config *Config) *Plugin {
	return &Plugin{newEnvironment(build, config)}
}

func (p *Plugin) Exec() error {
	if err := p.env.setUp(); err != nil {
		return err
	}
	defer p.env.tearDown()

	var cmds []*exec.Cmd
	cmds = append(cmds, commandVersion())
	cmds = append(cmds, p.env.commands()...)

	// Run ansible
	// execute all commands in batch mode.
	for _, cmd := range cmds {
		if err := runCommand(cmd); err != nil {
			return fmt.Errorf("failed to execute command: %v", err)
		}
	}

	return nil
}

// runCommand runs a command
func runCommand(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// helper function to create the docker info command.
func commandVersion() *exec.Cmd {
	return exec.Command(ansibleBin, "--version")
}

// Trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}
