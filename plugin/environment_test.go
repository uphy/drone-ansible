package plugin

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestEnvironmentSetup(t *testing.T) {
	e := newEnvironment(&Build{}, &Config{
		SSHKey:        "key",
		SSHPassphrase: "passphrase",
	})
	if err := e.setUp(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer e.tearDown()

	if e.files.ansibleCfg == "" {
		t.Errorf("ansible.cfg cannot be created")
	}
	if e.files.idrsa == "" {
		t.Errorf("id_rsa cannot be created")
	}
	if e.files.sshConfig == "" {
		t.Errorf("ssh_config cannot be created")
	}
}

func TestEnvironmentTearDown(t *testing.T) {
	e := newEnvironment(&Build{}, &Config{
		SSHKey:        "key",
		SSHPassphrase: "passphrase",
	})
	if err := e.setUp(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	if err := e.tearDown(); err != nil {
		t.Errorf("failed on tearDown: %v", err)
	}
	if _, err := os.Stat(e.files.ansibleCfg); !os.IsNotExist(err) {
		t.Error("file not deleted")
	}
}

func TestEnvironmentGenerateScript(t *testing.T) {
	testEnvironmentGenerateScript(t, Config{
		SSHUser:       "user1",
		SSHKey:        "key",
		SSHPassphrase: "passphrase",
		Playbook:      "playbook.yml",
	}, `#!/bin/sh -u
eval $(ssh-agent) > /dev/null || exit $?
echo "$SSH_PASSPHRASE" | SSH_ASKPASS="tmp/askpass.sh" ssh-add "tmp/id_rsa" > /dev/null 2>&1 || exit $?
/usr/bin/ansible-playbook -e "$1" -u "user1" -i "$2" "path/playbook.yml" 
ssh-agent -k > /dev/null || exit $?
`)
	testEnvironmentGenerateScript(t, Config{
		SSHKey:        "key",
		SSHPassphrase: "passphrase",
		Playbook:      "playbook.yml",
	}, `#!/bin/sh -u
eval $(ssh-agent) > /dev/null || exit $?
echo "$SSH_PASSPHRASE" | SSH_ASKPASS="tmp/askpass.sh" ssh-add "tmp/id_rsa" > /dev/null 2>&1 || exit $?
/usr/bin/ansible-playbook -e "$1" -i "$2" "path/playbook.yml" 
ssh-agent -k > /dev/null || exit $?
`)
	testEnvironmentGenerateScript(t, Config{
		Playbook: "playbook.yml",
	}, `#!/bin/sh -u
/usr/bin/ansible-playbook -e "$1" -i "$2" "path/playbook.yml" 
`)
}

func testEnvironmentGenerateScript(t *testing.T, config Config, expected string) {
	e := newEnvironment(&Build{
		Path: "path",
	}, &config)
	if err := e.setUp(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer e.tearDown()
	if s := generateScript(e); s != expected {
		t.Errorf("want %s, but %s", expected, s)
	}
}

func generateScript(e *environment) string {
	return strings.Replace(e.generateScript(), e.tmpDir, "tmp", -1)
}

func TestEnvironmentCommand(t *testing.T) {
	e := newEnvironment(&Build{
		Tag: "tag",
		SHA: "sha",
	}, &Config{
		SSHKey: "key",
	})
	if err := e.setUp(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer e.tearDown()
	cmd := e.command("inventory")
	args := cmd.Args
	if len(args) != 3 {
		t.Errorf("want %d but %d", 3, len(args))
	}
	if args[0] != e.files.script {
		t.Errorf("want %s but %s", e.files.script, args[0])
	}
	if extraVars := e.extraVars(); args[1] != extraVars {
		t.Errorf("want %s but %s", extraVars, args[1])
	}
	if args[2] != "inventory" {
		t.Errorf("want %s but %s", "inventory", args[2])
	}
	passphraseSet := false
	ansibleConfigSet := false
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "SSH_PASSPHRASE=") {
			passphraseSet = true
		}
		if strings.HasPrefix(env, "ANSIBLE_CONFIG=") {
			ansibleConfigSet = true
		}
	}
	if !passphraseSet {
		t.Error("SSH_PASSPHRASE not set.")
	}
	if !ansibleConfigSet {
		t.Error("ANSIBLE_CONFIG not set.")
	}
}

func TestEnvironmentExtraVars(t *testing.T) {
	testEnvironmentExtraVars(t, "key", map[string]string{
		"commit_tag":                   "tag",
		"commit_sha":                   "sha",
		"ansible_ssh_private_key_file": "/id_rsa",
	})
	testEnvironmentExtraVars(t, "", map[string]string{
		"commit_tag": "tag",
		"commit_sha": "sha",
	})
	testEnvironmentExtraVars(t, "", map[string]string{
		"commit_tag": "tag",
	})
	testEnvironmentExtraVars(t, "", map[string]string{
		"commit_sha": "sha",
	})
}

func testEnvironmentExtraVars(t *testing.T, key string, vars map[string]string) {
	e := newEnvironment(&Build{
		Tag: vars["commit_tag"],
		SHA: vars["commit_sha"],
	}, &Config{
		SSHKey: key,
	})
	if err := e.setUp(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer e.tearDown()
	actual := e.extraVars()
	actual = strings.Replace(actual, e.tmpDir, "", -1)
	var actualJSON map[string]string
	if err := json.Unmarshal([]byte(actual), &actualJSON); err != nil {
		t.Errorf("could'nt unmarshal json: %v", err)
	}
	if !reflect.DeepEqual(actualJSON, vars) {
		t.Errorf("want %#v, but %#v", vars, actualJSON)
	}
}
