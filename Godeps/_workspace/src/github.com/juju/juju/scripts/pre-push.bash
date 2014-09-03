#!/bin/bash

set -e

BADFMT=`find * -name '*.go' -not -name '.#*' | xargs gofmt -l`
if [ -n "$BADFMT" ]; then
	BADFMT=`echo "$BADFMT" | sed "s/^/  /"`
	echo -e "gofmt is sad:\n\n$BADFMT"
	exit 1
fi

VERSION=`go version | awk '{print $3}'`
go tool vet \
	-methods \
	-printf \
	-rangeloops \
	-printfuncs 'ErrorContextf:1,notFoundf:0,badReqErrorf:0,Commitf:0,Snapshotf:0,Debugf:0,Infof:0,Warningf:0,Errorf:0,Criticalf:0,Tracef:0' \
	.

# check this branch builds cleanly
go build github.com/juju/juju/...

# check that all tests are wired up
./scripts/checktesting.bash
