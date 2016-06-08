#!/bin/bash
set -x

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
MAKEDIR=$SCRIPTDIR/../../sidecar/

make -C $MAKEDIR build
STATUS=$?
if [ $STATUS -ne 0 ]; then
    echo -e "\n***********\nFAILED: make failed for sidecar.\n***********\n"
    exit $STATUS
fi

make -C $MAKEDIR docker IMAGE_NAME=a8-sidecar:0.1 DOCKERFILE=./docker/Dockerfile.ubuntu
STATUS=$?
if [ $STATUS -ne 0 ]; then
    echo -e "\n***********\nFAILED: docker build failed for sidecar (ubuntu version)\n***********\n"
    exit $STATUS
fi
