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

package commands

import (
	"fmt"

	"github.com/juju/loggo"
	"github.com/juju/cmd/envcmd"
	"github.com/juju/juju/instance"
	"github.com/juju/juju"
	"github.com/juju/juju/state"
	"github.com/juju/utils/ssh"
)

var logger = loggo.GetLogger("juju.sos.cmd")

type SosCommand struct {
	envcmd.EnvCommandBase
	Conn       *juju.Conn
	MachineMap map[string]*state.Machine
}

func (c *SosCommand) Query(target string) error {
	var err error
	c.Conn, err = juju.NewConnFromName(c.EnvName)
	if err != nil {
		return fmt.Errorf("Unable to connect to environment %q: %v", c.EnvName, err)
	}
	defer c.Conn.Close()

	c.MachineMap = make(map[string]*state.Machine)
	st := c.Conn.State

	if target == "" {
		logger.Infof("Querying all machines")

		machines, err := st.AllMachines()
		for _, m := range machines {
			// dont care about machine 0
			if m.Id() != "0" {
				logger.Infof("Adding machine(%s)", m.Id())
				c.MachineMap[m.Id()] = m
			}
		}
		if err != nil {
			return err
		}
		return nil
	}

	if target != "" {
		logger.Infof("Querying one machine(%s)", target)
		m, err := st.Machine(target)
		if err != nil {
			return fmt.Errorf("Unable to use machine(%s)", target)
		}
		c.MachineMap[m.Id()] = m
		return nil
	}
	return nil
}

func (c *SosCommand) ExecSsh(m *state.Machine) error {
	host := instance.SelectPublicAddress(m.Addresses())
	if host == "" {
		return fmt.Errorf("could not resolve machine's public address")
	}
	// make sure sosreport is installed
	// TODO: Remove when LP: #1311274 is released
	logger.Infof("Capturing sosreport for machine %s", m.Id())
	var options ssh.Options
	cmdStr := []string{"sudo apt-get install -yy sosreport && sudo sosreport --batch && sudo chown ubuntu:ubuntu -R /tmp/sosreport*"}
	cmd := ssh.Command("ubuntu@"+host, cmdStr, &options)
	return cmd.Run()
}
