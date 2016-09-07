#!/bin/bash

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source $SCRIPTDIR/.bluemixrc

echo "Listing existing container groups"
EXISTING_GROUPS=$(bluemix ic groups)

CONTROLPLANE_GROUPS=(
    amalgam8_controller
    amalgam8_registry
)

for group in ${CONTROLPLANE_GROUPS[@]}; do
    echo $EXISTING_GROUPS | grep $group > /dev/null
    if [ $? -eq 0 ]; then
        echo "Removing $group container group"
        bluemix ic group-remove $group
    fi
done

echo "Waiting for groups to be removed"
sleep 15
