// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package main

import (
	"errors"

	"github.com/juju/cmd"
	"launchpad.net/gnuflag"

	"github.com/juju/juju/cmd/envcmd"
)

// GetCommand retrieves the configuration of a service.
type GetCommand struct {
	envcmd.EnvCommandBase
	ServiceName string
	out         cmd.Output
}

func (c *GetCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "get",
		Args:    "<service>",
		Purpose: "get service configuration options",
	}
}

func (c *GetCommand) SetFlags(f *gnuflag.FlagSet) {
	// TODO(dfc) add json formatting ?
	c.out.AddFlags(f, "yaml", map[string]cmd.Formatter{
		"yaml": cmd.FormatYaml,
	})
}

func (c *GetCommand) Init(args []string) error {
	// TODO(dfc) add --schema-only
	if len(args) == 0 {
		return errors.New("no service name specified")
	}
	c.ServiceName = args[0]
	return cmd.CheckEmpty(args[1:])
}

// Run fetches the configuration of the service and formats
// the result as a YAML string.
func (c *GetCommand) Run(ctx *cmd.Context) error {
	client, err := c.NewAPIClient()
	if err != nil {
		return err
	}
	defer client.Close()

	results, err := client.ServiceGet(c.ServiceName)
	if err != nil {
		return err
	}

	resultsMap := map[string]interface{}{
		"service":  results.Service,
		"charm":    results.Charm,
		"settings": results.Config,
	}
	return c.out.Write(ctx, resultsMap)
}
