// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package mongo

var (
	MakeJournalDirs = makeJournalDirs
	MongoConfigPath = &mongoConfigPath
	NoauthCommand   = noauthCommand
	ProcessSignal   = &processSignal

	SharedSecretPath = sharedSecretPath
	SSLKeyPath       = sslKeyPath

	UpstartConfInstall          = &upstartConfInstall
	UpstartService              = upstartService
	UpstartServiceExists        = &upstartServiceExists
	UpstartServiceRunning       = &upstartServiceRunning
	UpstartServiceStopAndRemove = &upstartServiceStopAndRemove
	UpstartServiceStop          = &upstartServiceStop
	UpstartServiceStart         = &upstartServiceStart

	HostWordSize   = &hostWordSize
	RuntimeGOOS    = &runtimeGOOS
	AvailSpace     = &availSpace
	MinOplogSizeMB = &minOplogSizeMB
	MaxOplogSizeMB = &maxOplogSizeMB
	PreallocFile   = &preallocFile

	DefaultOplogSize  = defaultOplogSize
	FsAvailSpace      = fsAvailSpace
	PreallocFileSizes = preallocFileSizes
	PreallocFiles     = preallocFiles
)
