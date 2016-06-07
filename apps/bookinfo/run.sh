#!/bin/bash
set -x

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

$SCRIPTDIR/build-services.sh
if [ $? -ne 0 ]; then
    echo -e "\n***********\nFAILED: build-services.sh \n***********\n"
    exit 1
fi
$SCRIPTDIR/run-services.sh
if [ $? -ne 0 ]; then
    echo -e "\n***********\nFAILED: run-services.sh \n***********\n"
    exit 1
fi