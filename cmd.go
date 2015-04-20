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
	"fmt"

//	"os"
	"github.com/juju/loggo"
//	"github.com/juju/cmd"
	"github.com/juju/juju/cmd/envcmd"
	//"github.com/juju/juju/instance"
	"github.com/juju/juju/state"
	//"github.com/juju/juju/utils/ssh"
)

var logger = loggo.GetLogger("juju.sos.cmd")

type SosCommand struct {
	envcmd.EnvCommandBase
	state *state.State
	MachineMap map[string]*state.Machine
}

func (c *SosCommand) Query(target string) error {
	var err error
	client, err := c.NewAPIClient()

	if err != nil {
		return err
	}
	defer client.Close()

	c.MachineMap = make(map[string]*state.Machine)
	if target == "" {
		logger.Infof("Querying all machines")

		machines, err := c.state.AllMachines()
		for _, m := range machines {
			// dont care about machine 0
			if m.Id() != "0" {
				fmt.Println("Running")
				logger.Infof("Adding machine(%s)", m.Id())
				c.MachineMap[m.Id()] = m
			}
		}
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// func (c *SosCommand) ExecSsh(m *state.Machine) error {
// 	host := instance.SelectPublicAddress(m.Addresses())
// 	if host == "" {
// 		return fmt.Errorf("could not resolve machine's public address")
// 	}
// 	logger.Infof("Capturing sosreport for machine %s", m.Id())
// 	var options ssh.Options
// 	cmdStr := []string{"sudo sosreport --batch && sudo chown ubuntu:ubuntu -R /tmp/sosreport*"}
// 	cmd := ssh.Command("ubuntu@"+host, cmdStr, &options)
// 	return cmd.Run()
// }
func main() {
	fmt.Println("Running JUJU!")
	loggo.ConfigureLoggers("<root>=INFO")
	c := &SosCommand{}
//	cmd.Main(c, ctx, os.Args[1:])
	c.Query("")
}
