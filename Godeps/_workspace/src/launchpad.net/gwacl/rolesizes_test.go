// Copyright 2013 Canonical Ltd.  This software is licensed under the
// GNU Lesser General Public License version 3 (see the file COPYING).

package gwacl

import (
    . "launchpad.net/gocheck"
)

type rolesizeSuite struct{}

var _ = Suite(&rolesizeSuite{})

var knownRegions = []string{
    "East US",
    "West US",
    "North Central US",
    "South Central US",
    "North Europe",
    "West Europe",
    "Southeast Asia",
    "East Asia",
    "Japan East",
    "Japan West",
    "Brazil South",
}

var knownSizes = []string{
    "Basic_A0", "Basic_A1", "Basic_A2", "Basic_A3", "Basic_A4",
    "ExtraSmall", "Small", "Medium", "Large", "ExtraLarge",
    "A5", "A6", "A7", "A8", "A9",
}

func (suite *rolesizeSuite) TestRoleCostKnownRegions(c *C) {
    for _, region := range knownRegions {
        for _, roleSize := range knownSizes {
            cost, err := RoleSizeCost(region, roleSize)
            c.Check(err, IsNil)
            c.Check(cost, Not(Equals), uint64(0))
        }
    }
}

func (suite *rolesizeSuite) TestRoleCostUnknownRegion(c *C) {
    _, err := RoleSizeCost("Eastasia", "A0")
    c.Assert(err, ErrorMatches, `no cost data for region "Eastasia"`)
}

func (suite *rolesizeSuite) TestRoleCostUnknownRoleSize(c *C) {
    _, err := RoleSizeCost("East US", "A10")
    c.Assert(err, ErrorMatches, `no cost data for role size "A10" in region "East US"`)
}
