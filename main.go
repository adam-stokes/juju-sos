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
	"github.com/juju/loggo"
	"github.com/spf13/cobra"
	"runtime"
)

var logger = loggo.GetLogger("juju.plugin.sos")
var Destination string

var SosCmd = &cobra.Command{Use: "juju-sos",
	Short: "juju-sos is a juju plugin for capturing sosreport data",
	Long: `Capture sosreport data from multiple machines
or a single machine in a juju environment`,
	Run: capture,
}

func init() {
	SosCmd.Flags().StringVarP(&Destination, "destination", "d", "", "Output directory to store sos archives")
}

func capture(cmd *cobra.Command, args []string) {
	if Destination != "" {
		logger.Infof("Capturing and saving reports in: %s\n", Destination)
	} else {
		panic("A destination is required.")
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	loggo.ConfigureLoggers("<root>=INFO")
	logger.Infof("start this ish up.")

	SosCmd.Execute()
}
