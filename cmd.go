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
	"runtime"

	"github.com/juju/loggo"

	"launchpad.net/juju-core/cmd/envcmd"
	"launchpad.net/juju-core/instance"
	"launchpad.net/juju-core/juju"
	"launchpad.net/juju-core/names"
	"launchpad.net/juju-core/state"
	"launchpad.net/juju-core/utils/ssh"
)

var logger = loggo.GetLogger("juju.sos.cmd")

type SosCommand struct {
	envcmd.EnvCommandBase
	Conn *juju.Conn
	MachineMap map[string]*state.Machine
}

func (c *SosCommand) Connect(target string) error {
	var err error
	c.Conn, err = juju.NewConnFromName(c.EnvName)
	if err != nil {
		return fmt.Errorf("Unable to connect to environment %q: %v", c.EnvName, err)
	}
	defer c.Conn.Close()

	if !names.IsMachine(target) {
		return fmt.Errorf("invalid target: %q", target)
	}

	c.MachineMap = make(map[string]*state.Machine)
	st := c.Conn.State

	machines, err := st.AllMachines()
	for _, m := range machines {
		c.MachineMap[m.Id()] = m
	}

	return nil
}

func (c *SosCommand) ExecSsh(m *state.Machine) error {
	host := instance.SelectPublicAddress(m.Addresses())
	if host == "" {
		return fmt.Errorf("could not resolve machine's public address")
	}
	log.Println("Capturing sosreport for machine ", m.Id())
	var options ssh.Options
	cmd := ssh.Command("ubuntu@"+host, []string{"sudo sh -c soseport -b"}, &options)
	cmd.Stdin = strings.NewReader(script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
