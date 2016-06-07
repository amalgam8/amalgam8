#!/bin/bash
set -x

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

#################################################################################
# Build the helloworld image
#################################################################################
docker build -t hello:vx $SCRIPTDIR
