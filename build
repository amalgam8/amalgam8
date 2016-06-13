#!/bin/bash
set -x

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
STATUS=0

pushd $SCRIPTDIR
GO15VENDOREXPERIMENT=1 go build -a -o $SCRIPTDIR/docker/sidecar
STATUS=$?
popd

if [ $STATUS -ne 0 ]; then 
    echo "Compilation failed"
    exit $STATUS
fi

docker build -t kube-sidecar-proxy:0.1 -f $SCRIPTDIR/docker/Dockerfile.kube.proxy $SCRIPTDIR/docker
if [ $? -ne 0 ]; then
    echo -e "\n***********\nFAILED: docker build Dockerfile.kube.proxy failed.\n***********\n"
    exit 1
fi

docker build -t kube-sidecar-reg:0.1 -f $SCRIPTDIR/docker/Dockerfile.kube.reg $SCRIPTDIR/docker
if [ $? -ne 0 ]; then
    echo -e "\n***********\nFAILED: docker build Dockerfile.kube.reg failed.\n***********\n"
    exit 1
fi