package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	// without user
	if command := strings.Join(command(Build{
		Path: "path/to/playbookdir",
		SHA:  "123456789ABCDEF",
		Tag:  "TAGNAME",
	}, Config{
		InventoryPath: "inventory-path",
		Inventories:   []string{"inventory1", "inventory2"},
		Playbook:      "playbook.yml",
		SSHKey:        "",
	}, "inventory").Args, " "); command != `/usr/bin/ansible-playbook -e {"commit_sha":"123456789ABCDEF","commit_tag":"TAGNAME"} -i path/to/playbookdir/inventory-path/inventory path/to/playbookdir/playbook.yml` {
		t.Error("unexpected command: ", command)
	}

	// with user
	if command := strings.Join(command(Build{
		Path: "path/to/playbookdir",
		SHA:  "123456789ABCDEF",
		Tag:  "TAGNAME",
	}, Config{
		InventoryPath: "inventory-path",
		Inventories:   []string{"inventory1", "inventory2"},
		Playbook:      "playbook.yml",
		SSHUser:       "user1",
		SSHKey:        "",
	}, "inventory").Args, " "); command != `/usr/bin/ansible-playbook -e {"commit_sha":"123456789ABCDEF","commit_tag":"TAGNAME"} -u user1 -i path/to/playbookdir/inventory-path/inventory path/to/playbookdir/playbook.yml` {
		t.Error("unexpected command: ", command)
	}
}

func TestCommandEnvVars(t *testing.T) {
	testCommandEnvVars(t, map[string]string{
		"commit_tag":                   "tag",
		"commit_sha":                   "sha",
		"ansible_ssh_private_key_file": "/root/.ssh/id_rsa",
	})
	testCommandEnvVars(t, map[string]string{
		"commit_tag": "tag",
		"commit_sha": "sha",
	})
	testCommandEnvVars(t, map[string]string{
		"commit_tag": "tag",
	})
	testCommandEnvVars(t, map[string]string{
		"commit_sha": "sha",
	})
}

func testCommandEnvVars(t *testing.T, vars map[string]string) {
	actual := commandEnvVars(Build{
		Tag: vars["commit_tag"],
		SHA: vars["commit_sha"],
	}, Config{
		SSHKey: vars["ansible_ssh_private_key_file"],
	})
	if len(actual) != 2 {
		t.Fatalf("want 2 args, but %d", len(actual))
	}
	if actual[0] != "-e" {
		t.Errorf("want -e, but %s", actual[0])
	}
	var actualJSON map[string]string
	if err := json.Unmarshal([]byte(actual[1]), &actualJSON); err != nil {
		t.Errorf("could'nt unmarshal json: %v", err)
	}
	if !reflect.DeepEqual(actualJSON, vars) {
		t.Errorf("want %#v, but %#v", vars, actualJSON)
	}
}
