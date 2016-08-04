#!/bin/bash
set -x
set -o errexit
SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

git clone https://github.com/amalgam8/controller
git clone https://github.com/amalgam8/registry
git clone https://github.com/amalgam8/sidecar
git clone https://github.com/amalgam8/examples
sudo pip install git+https://github.com/amalgam8/a8ctl

$SCRIPTDIR/build-scripts/build-amalgam8.sh
$SCRIPTDIR/examples/docker/run-controlplane-docker.sh start
sleep 5
docker-compose -f $SCRIPTDIR/examples/docker/gateway.yaml up -d
sleep 5
docker-compose -f $SCRIPTDIR/examples/docker/bookinfo.yaml up -d
sleep 10
$SCRIPTDIR/testing/demo_script.sh
