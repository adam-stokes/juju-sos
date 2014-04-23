// juju-sos - Juju plugin for capturing sosreport data from the cloud
//
// Copyright (C) 2014 Adam Stokes <adam.stokes@ubuntu.com>
// Copyright (C) 2014 Canonical Ltd
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"os"
	"os/exec"
	"fmt"

	"github.com/juju/loggo"
	"launchpad.net/gnuflag"
	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/juju"

	// juju providers
	_ "launchpad.net/juju-core/provider/all"

	"github.com/battlemidget/juju-sos/commands"
)

var logger = loggo.GetLogger("juju.sos")

type SosCaptureCommand struct {
	commands.SosCommand
	target string
	destination string
}

var doc = `Capture sosreport data from multiple machines
or a single machine in a juju environment
`

func (c *SosCaptureCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name: "sos",
		Args: "[args] <target>",
		Purpose: "Capture sosreport from machine",
		Doc: doc,
	}
}

func (c *SosCaptureCommand) SetFlags(f *gnuflag.FlagSet) {
	c.SosCommand.SetFlags(f)
	f.StringVar(&c.destination, "d", "", "Output directory to store sos archives")
	f.StringVar(&c.target, "m", "", "(optional) Id of machine")
}

func (c *SosCaptureCommand) Init(args []string) error {
	err := c.SosCommand.Init()
	if err != nil {
		return err
	}
	if c.destination == "" {
		return fmt.Errorf("A destination is required, see `help` for more information.")
	}
	if c.destination != "" {
		finfo, err := os.Stat(c.destination)
		if err != nil {
			return fmt.Errorf("%q doesn't exist, you must create that directory first", c.destination)
		}
		if !finfo.IsDir() {
			return fmt.Errorf("Found %q, but it isn't a directory :(", c.destination)
		}
	}
	if c.target == "0" {
		return fmt.Errorf("Machine cannot be 0.")
	}
	return nil
}

func (c *SosCaptureCommand) Run(ctx *cmd.Context) error {
	var err error
	if c.target != "" {
		err = c.Query(c.target)
		if err != nil {
			return err
		}
		err = c.ExecSsh(c.MachineMap[c.target])
		if err != nil {
			return fmt.Errorf("Unable to run sosreport on machine: %s (%s)", c.target, err)
		}
		// scp
		logger.Infof("Copying archive to %q", c.destination)
		copyStr := exec.Command("juju","scp","--","-r", c.target+":/tmp/sosreport*xz", c.destination)
		copyStr.Stdout = os.Stdout
		err = copyStr.Run()
		if err != nil {
			return fmt.Errorf("Failed to copy sosreport: %v", err)
		}
	} else {
		// pass it a blank string :\
		err := c.Query("")
		if err != nil {
			return err
		}
		for _, m := range c.MachineMap {
			err = c.ExecSsh(m)
			if err != nil {
				// dont make this fatal
				logger.Errorf("Unable to run sosreport on machine: %d (%s)", m.Id(), err)
			}
			// scp
			logger.Infof("Copying archive to %q", c.destination)
			copyStr := exec.Command("juju","scp","--","-r", m.Id()+":/tmp/sosreport*xz", c.destination)
			copyStr.Stdout = os.Stdout
			copyStr.Stderr = os.Stderr
			err = copyStr.Run()
			if err != nil {
				return fmt.Errorf("Failed to copy sosreport: %v", err)
			}
		}
	}
	return nil
}

func main() {
	loggo.ConfigureLoggers("<root>=INFO")

	err := juju.InitJujuHome()
	if err != nil {
		panic(err)
	}
	ctx, err := cmd.DefaultContext()
	if err != nil {
		panic(err)
	}
	c := &SosCaptureCommand{}
	cmd.Main(c, ctx, os.Args[1:])
}
