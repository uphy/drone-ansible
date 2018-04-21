package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// environment has a responsibility for managing ansible environment.
type environment struct {
	build  *Build
	config *Config // Ansible config
	tmpDir string
	files  struct {
		ansibleCfg   string
		sshConfig    string
		idrsa        string
		dummyAskPass string
		script       string
	}
}

const (
	ansibleBin = "/usr/bin/ansible-playbook"
)

func newEnvironment(build *Build, config *Config) *environment {
	return &environment{
		build:  build,
		config: config,
	}
}

// setUp prepares ansible environment.
func (e *environment) setUp() error {
	// create temp dir for ansible environment
	dir, err := ioutil.TempDir("", "drone-ansible")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	e.tmpDir = dir

	// generate ssh_config
	if err := e.writeTempFile("ssh_config", "StrictHostKeyChecking no\nUserKnownHostsFile=/dev/null\n", 0600, &e.files.sshConfig, true); err != nil {
		return err
	}

	// generate ansible.cfg
	// this disables host key checking.. be aware of the man in the middle
	if err := e.writeTempFile("ansible.cfg", fmt.Sprintf(`[defaults]
host_key_checking = False
[ssh_connection]
ssh_args = -F "%s"
`, e.files.sshConfig), 0600, &e.files.ansibleCfg, true); err != nil {
		return err
	}

	// generate id_rsa
	if e.config.SSHKey != "" {
		if err := e.writeTempFile("id_rsa", e.config.SSHKey, 0600, &e.files.idrsa, false); err != nil {
			return err
		}
		if e.config.SSHPassphrase != "" {
			if err := e.writeTempFile("askpass.sh", "#!/bin/sh\nexec cat\n", 0700, &e.files.dummyAskPass, true); err != nil {
				return err
			}
		}
	}

	// generate script
	return e.writeTempFile("script.sh", e.generateScript(), 0700, &e.files.script, true)
}

func (e *environment) writeTempFile(name string, content string, perm os.FileMode, path *string, visible bool) error {
	*path = filepath.Join(e.tmpDir, name)
	if visible {
		e.dumpWithQuote(name, content)
	} else {
		e.dumpWithQuote(name, "<hidden>")
	}
	if err := ioutil.WriteFile(*path, []byte(content), perm); err != nil {
		return fmt.Errorf("failed to generate %s: %v", name, err)
	}
	return nil
}

func (e *environment) generateScript() string {
	script := "#!/bin/sh -u\n"
	if e.config.SSHKey != "" {
		script += "eval $(ssh-agent) > /dev/null || exit $?\n"
		script += fmt.Sprintf(`echo "$SSH_PASSPHRASE" | SSH_ASKPASS="%s" ssh-add "%s" > /dev/null 2>&1 || exit $?
`, e.files.dummyAskPass, e.files.idrsa)
	}
	ansibleCommand := []string{ansibleBin, "-e", `"$1"`}
	if e.config.SSHUser != "" {
		ansibleCommand = append(ansibleCommand, "-u", fmt.Sprintf(`"%s"`, e.config.SSHUser))
	}
	ansibleCommand = append(ansibleCommand, "-i", `"$2"`, fmt.Sprintf(`"%s"`, filepath.Join(e.build.Path, e.config.Playbook)), "\n")
	script += strings.Join(ansibleCommand, " ")
	if e.config.SSHKey != "" {
		script += "ssh-agent -k > /dev/null || exit $?\n"
	}
	return script
}

func (e *environment) dumpWithQuote(name, s string) error {
	if !e.config.Debug {
		return nil
	}
	fmt.Printf("[%s]\n", name)
	r := bufio.NewReader(strings.NewReader(s))
	for {
		l, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Println(">", string(l))
	}
	return nil
}

func (e *environment) commands() []*exec.Cmd {
	var cmds []*exec.Cmd
	for _, inventory := range e.config.Inventories {
		cmds = append(cmds, e.command(inventory))
	}
	return cmds
}

func (e *environment) command(inventory string) *exec.Cmd {
	cmd := exec.Command(e.files.script, e.extraVars(), filepath.Join(e.build.Path, inventory))
	cmd.Env = []string{
		fmt.Sprintf("SSH_PASSPHRASE=%s", e.config.SSHPassphrase),
		fmt.Sprintf("ANSIBLE_CONFIG=%s", e.files.ansibleCfg),
	}
	return cmd
}

func (e *environment) extraVars() string {
	vars := map[string]string{}
	if len(e.config.SSHKey) != 0 {
		vars["ansible_ssh_private_key_file"] = e.files.idrsa
	}
	if len(e.build.SHA) != 0 {
		vars["commit_sha"] = e.build.SHA
	}
	if len(e.build.Tag) != 0 {
		vars["commit_tag"] = e.build.Tag
	}

	b, err := json.Marshal(vars)
	if err != nil {
		panic(fmt.Sprint("unexpected err: ", err))
	}
	return string(b)
}

// tearDown clean the ansible environment.
func (e *environment) tearDown() error {
	if len(e.tmpDir) > 0 {
		if err := os.RemoveAll(e.tmpDir); err != nil {
			return fmt.Errorf("failed to delete temp dir: %v", e.tmpDir)
		}
	}
	return nil
}
