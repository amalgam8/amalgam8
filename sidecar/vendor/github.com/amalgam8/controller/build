#!/bin/bash
set -x

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
STATUS=0

pushd $SCRIPTDIR
GO15VENDOREXPERIMENT=1 go build -a -o $SCRIPTDIR/controller
STATUS=$?
popd

if [ $? -ne 0 ]; then
    echo -e "\n***********\nFAILED: go install failed for controller.\n***********\n"
    exit $STATUS
fi

docker build -t controller-0.1 -f $SCRIPTDIR/Dockerfile $SCRIPTDIR
if [ $? -ne 0 ]; then
    echo -e "\n***********\nFAILED: docker build failed for controller.\n***********\n"
    exit 1
fi