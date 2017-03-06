#!/bin/bash
#
# Copyright 2016 IBM Corporation
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

# docker or k8s
if [ -z "$A8_CONTAINER_ENV" ]; then
    A8_CONTAINER_ENV="docker"
fi
ENV=$A8_CONTAINER_ENV

# Dump some possibly useful data on failures to help debug
dump_debug() {

    if [ "$ENV" == "docker" ]; then
        curl localhost:31200/v1/rules
        docker exec -it gateway a8sidecar --debug show-state
        docker logs gateway
        docker exec gateway ps aux
    fi

    if [ "$ENV" == "k8s" ]; then
        kubectl get routingrule -o json
        PODNAME=$(kubectl get pods | grep gateway | awk '{print $1}')
        kubectl exec $PODNAME -- a8sidecar --debug show-state
        kubectl logs $PODNAME
        kubectl exec $PODNAME -- ps aux
    fi
}

MAX_LOOP=5

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
EXAMPLESDIR=$SCRIPTDIR/../../examples
A8CLI=$SCRIPTDIR/../../bin/a8ctl-linux

. $SCRIPTDIR/testing-common.sh

# On single runs, have seen the routing to not be 100% following the
# rule set.  So will loop to retry a few times to see if can get a
# passing test.
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
    ############# Set/Verify setting of default route #############

    # Clean up the routing rules from the previous testcase
    cleanup_all_rules

    echo ""
    echo "Set/verify the Hello World default route to v1"
    create_rule $EXAMPLESDIR/docker-helloworld-default-route-rules.json

    sleep 15

    echo ""
    echo "  Verifying the route was set to helloworld v1"
    temp_var1=$(list_rules "destination" 0)
    temp_var2=$(list_rules "route.backends[0].tags[0]" 0)

    if [ "${temp_var1}" != "helloworld" ] || [ "${temp_var2}" != "version=v1" ]; then
        echo "  The default route was not set as expected."
        echo "  Failed test"
        echo ""
        exit 1
    else
        echo "  Passed test"
        echo ""
    fi

    ############# Test traffic routes to v1 #############
    echo ""
    echo "Test Hello World traffic is routing 100% to v1"
    echo "  Call /helloworld/hello through the Gateway 100 times"

    v1_count=0
    for count in {1..100}
    do
        temp_var1=$(curl -s $A8_GATEWAY_URL/helloworld/hello)

        if [[ "${temp_var1}" = *"Hello version: version=v1"* ]]; then
            (( v1_count=v1_count+1 ))
        fi
    done
    echo "    v1 was hit: "$v1_count" times"

    if [ $v1_count -ne 100 ]; then
        # Test failed, routing was not 100%, try again
        echo "  The routing did not meet the rule that was set, try again."  
        (( retry_count=retry_count+1 ))
	dump_debug
    else
        # Test passed, time to exit the loop
        retry_count=100
    fi

    if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
        echo "  Test failed"
        echo ""
	dump_debug
        exit 1
    elif [ $retry_count -eq 100 ]; then
        echo "  Passed test"
        echo ""
    fi     
done

# On single runs, have seen the routing to not be 100% following the
# rule set.  So will loop to retry a few times to see if can get an
# approximate test of 75%/25% routing.
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
    ############# Split Traffic Between v1 and v2 #############

    # Clean up the routing rules from the previous testcase
    cleanup_all_rules

    echo ""
    echo "Set/verify the Hello World route to 75%/v1 25%/v2"
    create_rule $EXAMPLESDIR/docker-helloworld-v1-v2-route-rules.json

    sleep 20

    echo ""
    echo "  Verifying the rule was set."
    for count in {0..1}
    do
	temp_var1=$(list_rules "route.backends[0].tags[0]" $count)

        if [ "${temp_var1}" == "version=v1" ]; then
	    temp_var2=$(list_rules "route.backends[0].weight" $count)

            if [ "${temp_var2}" != "null" ]; then
                echo "    The v1 route was not set as the default 75% as expected."
                echo "  Failed test"
                echo ""
                exit 1
            else
                echo "    v1 route was set to handle 75% of the traffic as expected." 
            fi
        fi

        if [ "${temp_var1}" == "version=v2" ]; then
            temp_var2=$(list_rules "route.backends[0].weight" $count)

            if [ "${temp_var2}" != "0.25" ]; then
                echo "    The v2 route was not set to 25% as expected."
                echo "  Failed test"
                echo ""
                exit 1
            else
                echo "    v2 route was set to handle 25% of the traffic as expected." 
            fi
        fi
    done
    echo "  Passed test"
    echo ""

    ############# Test traffic is split between v1 and v2 #############
    echo ""
    echo "Test Hello World traffic is routing approximately 75%/v1 25%/v2"
    echo "  Call /helloworld/hello through the Gateway 100 times"

    v1_count=0
    v2_count=0
    for count in {1..100}
    do
        temp_var1=$(curl -s $A8_GATEWAY_URL/helloworld/hello) 

        if [[ "${temp_var1}" = *"Hello version: version=v1"* ]]; then
            (( v1_count=v1_count+1 ))
        elif [[ "${temp_var1}" = *"Hello version: version=v2"* ]]; then
            (( v2_count=v2_count+1 ))
        fi
    done
    echo "    v1 was hit: "$v1_count" times"
    echo "    v2 was hit: "$v2_count" times"
    echo ""

    if [ $v1_count -lt 70 ] || [  $v1_count -gt 80 ]; then
        # Test failed, routing outside 75%/25%, try again
        echo "  The routing did not meet the rule that was set, try again."  
        (( retry_count=retry_count+1 ))
	dump_debug
    else
        # Test passed, time to exit the loop
        retry_count=100
    fi

    if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
        echo "  Test failed"
        echo ""
        dump_debug
        exit 1
    elif [ $retry_count -eq 100 ]; then
        echo "  Passed test"
        echo ""
    fi     
done

# Clean up the routing rules from the previous testcase
cleanup_all_rules

echo ""
echo "Verification of the Hello World example application completed successfully."
echo ""
echo ""
