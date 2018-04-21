package main

import (
	"fmt"
	"os"

	"github.com/uphy/drone-ansible/plugin"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var build = "0" // build number set at compile-time

const (
	defaultPlaybook      = "provisioning/provision.yml"
	defaultInventoryPath = "provisioning/inventory"
)

func main() {
	app := cli.NewApp()
	app.Name = "ansible plugin"
	app.Usage = "ansible plugin"
	app.Action = run
	app.Version = fmt.Sprintf("1.0.%s", build)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "inventory-path",
			Usage:  "ansible inventory path",
			Value:  defaultInventoryPath,
			EnvVar: "PLUGIN_INVENTORY_PATH",
		},
		cli.StringSliceFlag{
			Name:   "inventories",
			Usage:  "ansible inventory files",
			Value:  &cli.StringSlice{"staging"},
			EnvVar: "PLUGIN_INVENTORY,PLUGIN_INVENTORIES",
		},
		cli.StringFlag{
			Name:   "playbook",
			Usage:  "ansible playbook to execute",
			Value:  defaultPlaybook,
			EnvVar: "PLUGIN_PLAYBOOK",
		},
		cli.StringFlag{
			Name:   "ssh-user",
			Usage:  "ssh user to access remote hosts",
			EnvVar: "SSH_USER,PLUGIN_SSH_USER",
		},
		cli.StringFlag{
			Name:   "ssh-key",
			Usage:  "ssh key to access remote hosts",
			EnvVar: "SSH_KEY,PLUGIN_SSH_KEY",
		},
		cli.StringFlag{
			Name:   "ssh-passphrase",
			Usage:  "ssh passphrase for the private key",
			EnvVar: "SSH_PASSPHRASE,PLUGIN_SSH_PASSPHRASE",
		},
		cli.StringFlag{
			Name:   "become-user",
			Usage:  "become(sudo) user name",
			EnvVar: "BECOME_USER,PLUGIN_BECOME_USER,SUDO_USER,PLUGIN_SUDO_USER",
		},
		cli.StringFlag{
			Name:   "become-password",
			Usage:  "become(sudo) password",
			EnvVar: "BECOME_PASSWORD,PLUGIN_BECOME_PASSWORD,SUDO_PASSWORD,PLUGIN_SUDO_PASSWORD",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug flag",
			EnvVar: "PLUGIN_DEBUG",
		},
		cli.StringFlag{
			Name:   "path",
			Usage:  "project base path",
			EnvVar: "DRONE_WORKSPACE",
		},
		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
		},
		cli.StringFlag{
			Name:   "commit.tag",
			Usage:  "project base path",
			EnvVar: "DRONE_TAG",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	return plugin.New(&plugin.Build{
		Path: c.String("path"),
		SHA:  c.String("commit.sha"),
		Tag:  c.String("commit.tag"),
	}, &plugin.Config{
		InventoryPath:  c.String("inventory-path"),
		Inventories:    c.StringSlice("inventories"),
		Playbook:       c.String("playbook"),
		SSHKey:         c.String("ssh-key"),
		SSHUser:        c.String("ssh-user"),
		SSHPassphrase:  c.String("ssh-passphrase"),
		BecomeUser:     c.String("become-user"),
		BecomePassword: c.String("become-password"),
		Debug:          c.Bool("debug"),
	}).Exec()
}
