// Copyright 2013-2014 Canonical Ltd.  This software is licensed under the
// GNU Lesser General Public License version 3 (see the file COPYING).

// Define the role sizes available in Azure.

package gwacl

import (
    "fmt"
)

// RoleSize is a representation of the machine specs available in the Azure
// documentation here:
//   http://msdn.microsoft.com/en-us/library/windowsazure/dn197896.aspx
//
// Pricing from here:
//   http://azure.microsoft.com/en-us/pricing/details/virtual-machines
//
// Detailed specifications here:
//   http://msdn.microsoft.com/en-us/library/windowsazure/dn197896.aspx
//
// Our specifications may be inaccurate or out of date.  When in doubt, check!
//
// The Disk Space values are only the maxumim permitted; actual space is
// determined by the OS image being used.
//
// Sizes and costs last updated 2014-06-23.
type RoleSize struct {
    Name          string
    CpuCores      uint64
    Mem           uint64 // In MB
    OSDiskSpace   uint64 // In MB
    TempDiskSpace uint64 // In MB
    MaxDataDisks  uint64 // 1TB each
}

// decicentsPerHour is the unit of cost we store for RoleSizeCost.
type decicentsPerHour uint64

const (
    // MB is the unit in which we specify sizes, so it's 1.
    // But please include it anyway, so that units are always explicit.
    MB  = 1
    GB  = 1024 * MB
    TB  = 1024 * GB
)

// Basic tier roles.
var basicRoleSizes = []RoleSize{{ // A0..A4: general purpose
    Name:          "Basic_A0",
    CpuCores:      1,  // shared
    Mem:           768 * MB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 20 * GB,
    MaxDataDisks:  1,
}, {
    Name:          "Basic_A1",
    CpuCores:      1,
    Mem:           1.75 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 40 * GB,
    MaxDataDisks:  2,
}, {
    Name:          "Basic_A2",
    CpuCores:      2,
    Mem:           3.5 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 60 * GB,
    MaxDataDisks:  4,
}, {
    Name:          "Basic_A3",
    CpuCores:      4,
    Mem:           7 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 120 * GB,
    MaxDataDisks:  8,
}, {
    Name:          "Basic_A4",
    CpuCores:      8,
    Mem:           14 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 240 * GB,
    MaxDataDisks:  16,
}}

// Standard tier roles.
var standardRoleSizes = []RoleSize{{ // A0..A4: general purpose
    Name:          "ExtraSmall", // A0
    CpuCores:      1,            // shared
    Mem:           768 * MB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 20 * GB,
    MaxDataDisks:  1,
}, {
    Name:          "Small", // A1
    CpuCores:      1,
    Mem:           1.75 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 70 * GB,
    MaxDataDisks:  2,
}, {
    Name:          "Medium", // A2
    CpuCores:      2,
    Mem:           3.5 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 135 * GB,
    MaxDataDisks:  4,
}, {
    Name:          "Large", // A3
    CpuCores:      4,
    Mem:           7 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 285 * GB,
    MaxDataDisks:  8,
}, {
    Name:          "ExtraLarge", // A4
    CpuCores:      8,
    Mem:           14 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 605 * GB,
    MaxDataDisks:  16,
}, { // A5..A7: memory intensive
    Name:          "A5",
    CpuCores:      2,
    Mem:           14 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 135 * GB,
    MaxDataDisks:  4,
}, {
    Name:          "A6",
    CpuCores:      4,
    Mem:           28 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 285 * GB,
    MaxDataDisks:  8,
}, {
    Name:          "A7",
    CpuCores:      8,
    Mem:           56 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 605 * GB,
    MaxDataDisks:  16,
}, { // A8..A9: compute intensive
    Name:          "A8",
    CpuCores:      8,
    Mem:           56 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 382 * GB,
    MaxDataDisks:  16,
}, {
    Name:          "A9",
    CpuCores:      16,
    Mem:           112 * GB,
    OSDiskSpace:   127 * GB,
    TempDiskSpace: 382 * GB,
    MaxDataDisks:  16,
}}

// RoleSizes describes all known role sizes.
var RoleSizes = append(append([]RoleSize{}, basicRoleSizes...), standardRoleSizes...)

var allRegionRoleCosts = map[string]map[string]decicentsPerHour{
    "East US": {
        "Basic_A0":   18,
        "Basic_A1":   44,
        "Basic_A2":   88,
        "Basic_A3":   176,
        "Basic_A4":   352,
        "ExtraSmall": 20,
        "Small":      60,
        "Medium":     120,
        "Large":      240,
        "ExtraLarge": 480,
        "A5":         250,
        "A6":         500,
        "A7":         1000,
        "A8":         1970,
        "A9":         4470,
    },
    "West US": {
        "Basic_A0":   18,
        "Basic_A1":   47,
        "Basic_A2":   94,
        "Basic_A3":   188,
        "Basic_A4":   376,
        "ExtraSmall": 20,
        "Small":      60,
        "Medium":     120,
        "Large":      240,
        "ExtraLarge": 480,
        "A5":         250,
        "A6":         500,
        "A7":         1000,
        "A8":         1970,
        "A9":         4470,
    },
    "North Central US": {
        "Basic_A0":   18,
        "Basic_A1":   47,
        "Basic_A2":   94,
        "Basic_A3":   188,
        "Basic_A4":   376,
        "ExtraSmall": 20,
        "Small":      60,
        "Medium":     120,
        "Large":      240,
        "ExtraLarge": 480,
        "A5":         250,
        "A6":         500,
        "A7":         1000,
        "A8":         1970,
        "A9":         4470,
    },
    "South Central US": {
        "Basic_A0":   18,
        "Basic_A1":   44,
        "Basic_A2":   88,
        "Basic_A3":   176,
        "Basic_A4":   352,
        "ExtraSmall": 20,
        "Small":      60,
        "Medium":     120,
        "Large":      240,
        "ExtraLarge": 480,
        "A5":         220,
        "A6":         440,
        "A7":         880,
        "A8":         1970,
        "A9":         4470,
    },
    "North Europe": {
        "Basic_A0":   18,
        "Basic_A1":   47,
        "Basic_A2":   94,
        "Basic_A3":   188,
        "Basic_A4":   376,
        "ExtraSmall": 20,
        "Small":      60,
        "Medium":     120,
        "Large":      240,
        "ExtraLarge": 480,
        "A5":         248,
        "A6":         496,
        "A7":         992,
        "A8":         1970,
        "A9":         4470,
    },
    "West Europe": {
        "Basic_A0":   18,
        "Basic_A1":   51,
        "Basic_A2":   102,
        "Basic_A3":   204,
        "Basic_A4":   408,
        "ExtraSmall": 20,
        "Small":      60,
        "Medium":     120,
        "Large":      240,
        "ExtraLarge": 480,
        "A5":         270,
        "A6":         540,
        "A7":         1080,
        "A8":         1970,
        "A9":         4470,
    },
    "Southeast Asia": {
        "Basic_A0":   18,
        "Basic_A1":   58,
        "Basic_A2":   116,
        "Basic_A3":   232,
        "Basic_A4":   464,
        "ExtraSmall": 20,
        "Small":      60,
        "Medium":     120,
        "Large":      240,
        "ExtraLarge": 480,
        "A5":         270,
        "A6":         540,
        "A7":         1080,
        "A8":         1970,
        "A9":         4470,
    },
    "East Asia": {
        "Basic_A0":   18,
        "Basic_A1":   58,
        "Basic_A2":   116,
        "Basic_A3":   232,
        "Basic_A4":   464,
        "ExtraSmall": 20,
        "Small":      60,
        "Medium":     120,
        "Large":      240,
        "ExtraLarge": 480,
        "A5":         294,
        "A6":         588,
        "A7":         1176,
        "A8":         1970,
        "A9":         4470,
    },
    "Japan East": {
        "Basic_A0":   18,
        "Basic_A1":   69,
        "Basic_A2":   138,
        "Basic_A3":   276,
        "Basic_A4":   552,
        "ExtraSmall": 27,
        "Small":      81,
        "Medium":     162,
        "Large":      324,
        "ExtraLarge": 648,
        "A5":         281,
        "A6":         562,
        "A7":         1124,
        "A8":         1970,
        "A9":         4470,
    },
    "Japan West": {
        "Basic_A0":   18,
        "Basic_A1":   61,
        "Basic_A2":   122,
        "Basic_A3":   244,
        "Basic_A4":   488,
        "ExtraSmall": 25,
        "Small":      73,
        "Medium":     146,
        "Large":      292,
        "ExtraLarge": 584,
        "A5":         258,
        "A6":         516,
        "A7":         1032,
        "A8":         1970,
        "A9":         4470,
    },
    "Brazil South": {
        "Basic_A0":   22,
        "Basic_A1":   58,
        "Basic_A2":   116,
        "Basic_A3":   232,
        "Basic_A4":   464,
        "ExtraSmall": 27,
        "Small":      80,
        "Medium":     160,
        "Large":      320,
        "ExtraLarge": 640,
        "A5":         291,
        "A6":         582,
        "A7":         1164,
        "A8":         1970,
        "A9":         4470,
    },
}

// RoleSizeCost returns the cost associated with the given role size and region.
func RoleSizeCost(region string, roleSize string) (decicentsPerHour uint64, err error) {
    costs, ok := allRegionRoleCosts[region]
    if !ok {
        return 0, fmt.Errorf("no cost data for region %q", region)
    }
    cost, ok := costs[roleSize]
    if ok {
        return uint64(cost), nil
    }
    return 0, fmt.Errorf(
        "no cost data for role size %q in region %q",
        roleSize, region,
    )
}
