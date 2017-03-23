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

# Set local vars
SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
EXAMPLESDIR=$SCRIPTDIR/../../examples
TCPAPPDIR=$SCRIPTDIR/../apps/tcp
A8_TEST_SUITE=$1
if [ "$A8_TEST_SUITE" == "examples" ]; then
    PRODUCTPAGE_V1_OUTPUT="productpage_v1.html"
    PRODUCTPAGE_V2_OUTPUT="productpage_v2.html"
    PRODUCTPAGE_V3_OUTPUT="productpage_v3.html"
    PRODUCTPAGE_RULEMATCH="productpage_rulematch.html"
    RULESDIR=$EXAMPLESDIR
else
    PRODUCTPAGE_V1_OUTPUT="productpage_v1.json"
    PRODUCTPAGE_V2_OUTPUT="productpage_v2.json"
    PRODUCTPAGE_V3_OUTPUT="productpage_v3.json"
    PRODUCTPAGE_RULEMATCH="productpage_rulematch.json"
    TCPHELLOWORLD_OUTPUT="tcphelloworld_output"
    RULESDIR=$SCRIPTDIR
fi
# docker or k8s
if [ -z "$A8_CONTAINER_ENV" ]; then
    A8_CONTAINER_ENV="docker"
fi
ENV=$A8_CONTAINER_ENV

dump_debug() {
    if [ "$ENV" == "k8s" ]; then
        kubectl get pods
        kubectl get routingrule
        kubectl get routingrule -o json
        GATEWAY_PODNAME=$(kubectl get pods | grep gateway | awk '{print $1}')
        kubectl logs $GATEWAY_PODNAME
        kubectl exec -it $GATEWAY_PODNAME cat /etc/envoy/envoy.json
        kubectl exec -it $GATEWAY_PODNAME tail /var/log/a8_access.log
        kubectl exec -it $GATEWAY_PODNAME curl localhost:6500/v1/clusters/blah/blah
        kubectl exec -it $GATEWAY_PODNAME curl localhost:6500/v1/routes/blah/blah/blah
        PRODUCTPAGE_PODNAME=$(kubectl get pods | grep productpage | awk '{print $1}')
        kubectl logs $PRODUCTPAGE_PODNAME -c productpage
        kubectl logs $PRODUCTPAGE_PODNAME -c serviceproxy
        RULES_PODNAME=$(kubectl get pods | grep rules | awk '{print $1}')
        kubectl logs $RULES_PODNAME
    fi
}

jsondiff() {
    # Sort JSON fields, or fallback to original text
    file1=$(jq -S . $1 2> /dev/null) || file1=$(cat $1)
    file2=$(jq -S . $2 2> /dev/null) || file2=$(cat $2)

    # Diff, but ignore all whitespace since
    diff -Ewb <(echo $file1) <(echo $file2) &>/dev/null
}

check_output_equality() {
    REF_OUTPUT_FILE="$1"
    ACTUAL_OUTPUT_FILE="$2"
    if [ "$A8_TEST_SUITE" == "examples" ]; then
	    diff $REF_OUTPUT_FILE $ACTUAL_OUTPUT_FILE > /dev/null
    else
	    jsondiff $REF_OUTPUT_FILE $ACTUAL_OUTPUT_FILE > /dev/null
    fi

    if [ $? -gt 0 ]; then
        return 1
    fi
    return 0
}

# Call the specified endpoint and compare against expected output
# Ensure the % falls within the expected range
check_routing_rules() {
    COMMAND_INPUT="$1"
    EXPECTED_OUTPUT1="$2"
    EXPECTED_OUTPUT2="$3"
    EXPECTED_PERCENT="$4"
    MAX_LOOP=5
    routing_retry_count=1
    COMMAND_INPUT="${COMMAND_INPUT} >/tmp/routing.tmp"

    while [  $routing_retry_count -le $((MAX_LOOP)) ]; do
        v1_count=0
        v2_count=0
        for count in {1..100}
        do
            temp_var1=$(eval $COMMAND_INPUT)
            check_output_equality $EXPECTED_OUTPUT1 "/tmp/routing.tmp"
            if [ $? -eq 0 ]; then
                (( v1_count=v1_count+1 ))
            else
                check_output_equality $EXPECTED_OUTPUT2 "/tmp/routing.tmp"
                if [ $? -eq 0 ]; then
                    (( v2_count=v2_count+1 ))
                fi
            fi
        done
        echo "    v1 was hit: "$v1_count" times"
        echo "    v2 was hit: "$v2_count" times"
        echo ""

        EXPECTED_V1_PERCENT=$((100-$EXPECTED_PERCENT))
        ADJUST=5
        if [ $v1_count -lt $(($EXPECTED_V1_PERCENT-$ADJUST)) ] || [  $v1_count -gt $(($EXPECTED_V1_PERCENT+$ADJUST)) ]; then
            echo "  The routing did not meet the rule that was set, try again."
            (( routing_retry_count=routing_retry_count+1 ))
        else
            # Test passed, time to exit the loop
            routing_retry_count=100
        fi

        if [ $routing_retry_count -eq $((MAX_LOOP+1)) ]; then
            echo "  Test failed"
            echo ""
            return 1
        elif [ $routing_retry_count -eq 100 ]; then
            echo "  Passed test"
            echo ""
        fi
    done
    return 0
}

A8CLI=$SCRIPTDIR/../../bin/a8ctl-linux

. $SCRIPTDIR/testing-common.sh

cleanup_all_rules
sleep 30

######Verify services have default route#######
echo "testing default route behavior.."
MAX_LOOP=5
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
    response=$(curl --write-out %{http_code} --silent --output /dev/null http://localhost:32000/productpage/productpage)
    if [ $response -ne 200 ]; then
        (( retry_count=retry_count+1 ))
        if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
            exit 1
        fi
        echo "failed, retrying"
        echo "Prior to creating any rules, services are unreachable."
        sleep 10
    else
        break
    fi
done
echo "works!"


############Test TCP Bridge in Envoy##############
# Only supported in Docker environment currently
if [ "$ENV" == "docker" ]; then
    echo "testing tcp bridge in envoy..."
    retry_count=1
    while [  $retry_count -le $((MAX_LOOP)) ]; do
        echo  "Sending a message to tcphelloworld, expecting same words echoing back..."
        python $TCPAPPDIR/test.py localhost 12345 > /tmp/$TCPHELLOWORLD_OUTPUT
        diff $SCRIPTDIR/$TCPHELLOWORLD_OUTPUT /tmp/$TCPHELLOWORL_OUTPUT

        if [ $? -gt 0 ]; then
            echo "failed."
	    echo "The message received does not match the message sent to service"
	    exit 1
        else
	    break
        fi
    done
    echo "works!"
fi
######Version Routing##############
echo "testing version routing.."
MAX_LOOP=5
retry_count=1

create_rule $RULESDIR/bookinfo-default-route-details.yaml
create_rule $RULESDIR/bookinfo-default-route-productpage.yaml
create_rule $RULESDIR/bookinfo-default-route-ratings.yaml
create_rule $RULESDIR/bookinfo-default-route-reviews.yaml
create_rule $RULESDIR/bookinfo-jason-reviews-v2-route-rules.yaml

sleep 30

while [  $retry_count -le $((MAX_LOOP)) ]; do
    echo -n "injecting traffic for user=shriram, expecting productpage_v1.."
    curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/$PRODUCTPAGE_V1_OUTPUT

    if [ "$A8_TEST_SUITE" == "examples" ]; then
	    diff $SCRIPTDIR/$PRODUCTPAGE_V1_OUTPUT /tmp/$PRODUCTPAGE_V1_OUTPUT
    else
	    jsondiff $SCRIPTDIR/$PRODUCTPAGE_V1_OUTPUT /tmp/$PRODUCTPAGE_V1_OUTPUT
    fi

    if [ $? -gt 0 ]; then
        (( retry_count=retry_count+1 ))
        if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
            exit 1
        fi
        echo "failed, retrying"
        echo "Productpage not match productpage_v1 after setting route to reviews:v2 for user=shriram in version routing test phase"
	dump_debug
	sleep 10
    else
        break
    fi
done
echo "works!"

retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
    echo -n "injecting traffic for user=jason, expecting productpage_v2.."
    curl -s -b 'foo=bar;user=jason;ding=dong;x' http://localhost:32000/productpage/productpage >/tmp/$PRODUCTPAGE_V2_OUTPUT

    if [ "$A8_TEST_SUITE" == "examples" ]; then
	diff $SCRIPTDIR/$PRODUCTPAGE_V2_OUTPUT /tmp/$PRODUCTPAGE_V2_OUTPUT
    else
	jsondiff $SCRIPTDIR/$PRODUCTPAGE_V2_OUTPUT /tmp/$PRODUCTPAGE_V2_OUTPUT
    fi

    if [ $? -gt 0 ]; then
        (( retry_count=retry_count+1 ))
        if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
            exit 1
        fi
        echo "failed, retrying"
        echo "Productpage does not match productpage_v2 after setting route to reviews:v2 for user=jason in version routing test phase"
	sleep 10
    else
        break
    fi
done
echo "works!"

########Fault injection
echo "testing fault injection.."
output=$(create_rule $RULESDIR/bookinfo-jason-7s-delay.yaml)
fault_rule_id=$( tr -d ' \t\n\r' <<< "$output" | sed -e 's/\({"ids":\["\)//g' -e 's/\("\]}\)//g' )

sleep 20

retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
    ###For shriram
    echo -n "injecting traffic for user=shriram, expecting productpage_v1 in less than 2s.."
    before=$(date +"%s")
    curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/$PRODUCTPAGE_V1_OUTPUT
    after=$(date +"%s")

    delta=$(($after-$before))
    if [ $delta -gt 2 ]; then
        (( retry_count=retry_count+1 ))
        if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
            exit 1
        fi
        echo "failed, retrying"
        echo "Productpage took more $delta seconds to respond (expected 2s) for user=shriram in fault injection phase"
        continue
    fi

    if [ "$A8_TEST_SUITE" == "examples" ]; then
        diff $SCRIPTDIR/$PRODUCTPAGE_V1_OUTPUT /tmp/$PRODUCTPAGE_V1_OUTPUT
    else
        jsondiff $SCRIPTDIR/$PRODUCTPAGE_V1_OUTPUT /tmp/$PRODUCTPAGE_V1_OUTPUT
    fi

    if [ $? -gt 0 ]; then
        echo "failed"
        echo "Productpage does not match productpage_v1 after injecting fault rule for user=shriram in fault injection test phase"
        exit 1
    else
        break
    fi
done
echo "works!"

retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
    ####For jason
    echo -n "injecting traffic for user=jason, expecting 'reviews unavailable' message in approx 7s.."
    before=$(date +"%s")
    curl -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage >/tmp/$PRODUCTPAGE_RULEMATCH
    after=$(date +"%s")

    delta=$(($after-$before))
    if [ $delta -gt 8 -o $delta -lt 5 ]; then
        (( retry_count=retry_count+1 ))
        if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
            exit 1
        fi
        echo "failed, retrying"
        echo "Productpage took more $delta seconds to respond (expected <8s && >5s) for user=jason in fault injection phase"
        continue
    fi

    if [ "$A8_TEST_SUITE" == "examples" ]; then
        diff $SCRIPTDIR/$PRODUCTPAGE_RULEMATCH /tmp/$PRODUCTPAGE_RULEMATCH
    else
        jsondiff $SCRIPTDIR/$PRODUCTPAGE_RULEMATCH /tmp/$PRODUCTPAGE_RULEMATCH
    fi

    if [ $? -gt 0 ]; then
        echo "failed"
        echo "Productpage does not match productpage_rulematch.json after injecting fault rule for user=jason in fault injection test phase"
        exit 1
    else
        break
    fi
done
echo "works!"


#####Clear rules and check
echo "clearing fault injection rule.."
delete_rule $fault_rule_id
sleep 15

retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
    echo -n "Testing app again for user=jason, expecting productpage_v2 in less than 2s.."
    before=$(date +"%s")
    curl -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage >/tmp/$PRODUCTPAGE_V2_OUTPUT
    after=$(date +"%s")

    delta=$(($after-$before))
    if [ $delta -gt 2 ]; then
        (( retry_count=retry_count+1 ))
        if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
            exit 1
        fi
        echo "failed, retrying"
        echo "Productpage took more $delta seconds to respond (expected <2s) after clearing rules for user=jason in fault injection phase"
        continue
    fi

    if [ "$A8_TEST_SUITE" == "examples" ]; then
        diff $SCRIPTDIR/$PRODUCTPAGE_V2_OUTPUT /tmp/$PRODUCTPAGE_V2_OUTPUT
    else
        jsondiff $SCRIPTDIR/$PRODUCTPAGE_V2_OUTPUT /tmp/$PRODUCTPAGE_V2_OUTPUT
    fi

    if [ $? -gt 0 ]; then
        echo "failed"
        echo "Productpage does not match productpage_v2.json after clearing fault rules for user=jason in fault injection test phase"
        exit 1
    else
        break
    fi
done
echo "works!"

#######Gradually migrate traffic to reviews:v3 for all users
cleanup_all_rules

# Quiet things down a bit
set +x

MAX_LOOP=5
retry_count=1
SLEEP_TIME=15
GATEWAY_URL=localhost:32000
COMMAND_INPUT="curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage"
EXPECTED_OUTPUT1="$SCRIPTDIR/$PRODUCTPAGE_V1_OUTPUT"
EXPECTED_OUTPUT2="$SCRIPTDIR/$PRODUCTPAGE_V3_OUTPUT"

MAX_LOOP=5
retry_count=1
echo "Percentage based routing. 25% to v1 and 75% to v3."

create_rule $RULESDIR/bookinfo-reviews-routing.yaml

while [  $retry_count -le $((MAX_LOOP)) ]; do
	# Give it a bit to process the request
	sleep $SLEEP_TIME

	# Validate that 25% of traffic is routing to v1
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT2" "$EXPECTED_OUTPUT1" 25
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			echo "Exiting"
			exit 1
		fi
		echo "Retrying..."
	else
		break
	fi
done

echo "done"
