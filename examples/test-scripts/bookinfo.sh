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
ENV=$1

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

A8_CLI=$SCRIPTDIR/../../bin/a8ctl-beta

# Dump some possibly useful data on failures to help debug
dump_debug() {
    curl localhost:31200/v1/rules

    if [ "$ENV" == "docker" ]; then
        $A8_CLI service-list
        $A8_CLI route-list
        docker exec -it gateway a8sidecar --debug show-state
        docker logs gateway
        docker exec gateway ps aux
        docker exec gateway cat /var/log/nginx/access.log
        docker exec gateway cat /var/log/nginx/error.log
        PID=$(docker exec gateway cat /var/run/nginx.pid)
        echo $PID
        OUTPUT=$(docker exec gateway kill -s 0 $PID)
        echo $?
        echo $OUTPUT
    fi

    if [ "$ENV" == "k8s" ]; then
        PODNAME=$(kubectl get pods | grep gateway | awk '{print $1}')
        kubectl exec $PODNAME -- a8sidecar --debug show-state
        kubectl logs $PODNAME
        kubectl exec $PODNAME -- ps aux
        kubectl exec $PODNAME -- cat /var/log/nginx/access.log
        kubectl exec $PODNAME -- cat /var/log/nginx/error.log
    fi
}

jsondiff_test() {
	# Diff, but ignore all whitespace since
	diff -Ewb <(echo $1) <(echo $2)
}

jsondiff() {
	# Sort JSON fields, or fallback to original text
	file1=$(jq 'sort_by(.service)' $1 2> /dev/null) || file1=$(cat $1)
	file2=$(jq 'sort_by(.service)' $2 2> /dev/null) || file2=$(cat $2)

	# Diff, but ignore all whitespace since
	diff -Ewb <(echo $file1) <(echo $file2)
}

jsondiff_service_list() {
	# Sort JSON fields, or fallback to original text
	file1=$(jq 'sort_by(.service)' $1 | jq '.[].instances|=sort' 2> /dev/null) || file1=$(cat $1)
	file2=$(jq 'sort_by(.service)' $2 | jq '.[].instances|=sort' 2> /dev/null) || file2=$(cat $2)

	# Diff, but ignore all whitespace since
	diff -Ewb <(echo $file1) <(echo $file2)
}

jsondiff_action_list() {
	# Sort JSON fields, or fallback to original text
	# Remove the id field as that will change between runs
	file1=$(jq -S 'del(.[].id)' $1 2> /dev/null) || file1=$(cat $1)
	file2=$(jq -S 'del(.[].id)' $2 2> /dev/null) || file2=$(cat $2)

	# Diff, but ignore all whitespace since
	diff -Ewb <(echo $file1) <(echo $file2)
}

jsondiff_recipe_results() {
	# Sort JSON fields, or fallback to original text
        file1=$(jq -S '.' $1 2> /dev/null) || file1=$(cat $1)
        file2=$(jq -S '.' $2 2> /dev/null) || file2=$(cat $2)

        # Diff, but ignore all whitespace since
        diff -Ewb <(echo $file1) <(echo $file2)
}

# Call the specified endpoint and compare against expected output
# Ensure the % falls within the expected range
check_routing_rules() {
	COMMAND_INPUT="$1"
	EXPECTED_OUTPUT1="$2"
	EXPECTED_OUTPUT2="$3"
	EXPECTED_PERCENT="$4"
	MAX_LOOP=5
	retry_count=1
	while [  $retry_count -le $((MAX_LOOP)) ]; do
		v1_count=0
		v2_count=0
		for count in {1..100}
		do
			temp_var1=$(eval $COMMAND_INPUT)

			if [[ "${temp_var1}" = *"$EXPECTED_OUTPUT1"* ]]; then
				(( v1_count=v1_count+1 ))
			elif [[ "${temp_var1}" = *"$EXPECTED_OUTPUT2"* ]]; then
				(( v2_count=v2_count+1 ))
			fi
		done
		echo "    v1 was hit: "$v1_count" times"
		echo "    v2 was hit: "$v2_count" times"
		echo ""

		EXPECTED_V1_PERCENT=$((100-$EXPECTED_PERCENT))
		ADJUST=5
		if [ $v1_count -lt $(($EXPECTED_V1_PERCENT-$ADJUST)) ] || [  $v1_count -gt $(($EXPECTED_V1_PERCENT+$ADJUST)) ]; then
			echo "  The routing did not meet the rule that was set, try again."
			(( retry_count=retry_count+1 ))
		else
			# Test passed, time to exit the loop
			retry_count=100
		fi

		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			echo "  Test failed"
			echo ""
			dump_debug
			return 1
		elif [ $retry_count -eq 100 ]; then
			echo "  Passed test"
			echo ""
		fi
	done
	return 0
}

cleanup_and_exit() {
	$A8_CLI rule-delete -a -f
	dump_debug
	exit 1
}

# Param 1: rule type: route-list, action-list
# Param 2: rules input file name
# Param 3: rules expected output file name
create_rule() {
	$A8_CLI rule-create -f $2 > /dev/null
	sleep 5
	# route-list is only valid for docker-compose
	if [ "$ENV" == "docker" -a "$1" == "route-list" ]; then
		$A8_CLI $1 -o json > /tmp/rule_output.json
		jsondiff $3 /tmp/rule_output.json
		if [ $? -gt 0 ]; then
			echo "failed"
			echo "Actual a8ctl $1 output does not match the expected output."
			cleanup_and_exit
		fi
	fi

	if [ "$1" == "action-list" ]; then
		$A8_CLI $1 -o json > /tmp/rule_output.json
                jsondiff_action_list $3 /tmp/rule_output.json
                if [ $? -gt 0 ]; then
                        echo "failed"
                        echo "Actual a8ctl $1 output does not match the expected output."
                        cleanup_and_exit
                fi
	fi
}

echo "===== Starting Bookinfo demo ====="
echo "===== Step 0: Cleanup ====="

$A8_CLI rule-delete -a -f

echo "===== Step 1: List the services in the app ====="

# The cli service-list doesn't work when using the k8s registry
if [ "$ENV" == "docker" ]; then
	$A8_CLI service-list -o json > /tmp/service-list.json
	NUM_SERVICES=$(jq 'length' /tmp/service-list.json)
	COUNTER=0
	# Expect 4 services to be returned by the service-list call
	# They might not all be registered yet so retry a few times
	while [ $NUM_SERVICES -lt 4 ]; do
		if [ $COUNTER -gt 6 ]; then
			echo "Number of retries exceeded. Exiting."
			exit 1
		fi

		sleep 10
		$A8_CLI service-list -o json > /tmp/service-list.json
		NUM_SERVICES=$(jq 'length' /tmp/service-list.json)
		let COUNTER=COUNTER+1
	done

	jsondiff_service_list $SCRIPTDIR/output/service-list.json /tmp/service-list.json
	if [ $? -gt 0 ]; then
		echo "failed"
		echo "Actual a8ctl service-list output does not match the expected output."
		exit 1
	fi
fi
echo "Step 1 passed"

echo "===== Step 2: Set the default routes and verify v1 productpage ====="

MAX_LOOP=5
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
	create_rule route-list $SCRIPTDIR/rules/default-route-rules.json $SCRIPTDIR/output/default-route-list.json

	echo "injecting traffic for user=shriram, expecting productpage_v1.."
	curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/productpage_v1.html

	diff $SCRIPTDIR/output/productpage_v1.html /tmp/productpage_v1.html
	if [ $? -gt 0 ]; then
                (( retry_count=retry_count+1 ))
		dump_debug
		echo "failed"
		echo "Productpage not match productpage_v1 after setting route to reviews:v2 for user=shriram in version routing test phase"
                if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
                        cleanup_and_exit
                fi
		sleep 10
	else
		break
	fi
done
echo "Step 2 passed"

echo "===== Step 3: Enable ratings service for user jason ====="

create_rule route-list $SCRIPTDIR/rules/v2-selector-rule.json $SCRIPTDIR/output/route-list-v2.json

echo "Step 3 passed"

echo "===== Step 4: Simple fault injection ====="

MAX_LOOP=5
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
	create_rule action-list $SCRIPTDIR/rules/action-add-delay.json $SCRIPTDIR/output/action-list.json

	# If the user is not jason, we do not expect a delay
	NO_DELAY_LOAD_TIME=$(curl -w %{time_total} -o /dev/null -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage)
	if (( $(echo "$NO_DELAY_LOAD_TIME > 1" |bc -l) )); then
                (( retry_count=retry_count+1 ))
		dump_debug
		echo "failed"
		echo "Too much of a delay reading productpage"
                if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			cleanup_and_exit
		fi
		continue
	fi

	# If the user is jason, we expect a 6 second delay
	DELAY_LOAD_TIME=$(curl -w %{time_total} -o /dev/null -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage)
	if (( $(echo "$DELAY_LOAD_TIME < 6" |bc -l) )); then
                (( retry_count=retry_count+1 ))
		dump_debug
		echo "failed"
		echo "Not enough of a delay reading productpage: $DELAY_LOAD_TIME"
                if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			cleanup_and_exit
		fi
		echo "Retrying...."
	else
		break
	fi
done

echo "Step 4 passed"

# Cleanup the rules
$A8_CLI rule-delete -a -f

echo "===== Step 5: Fault injection + automated verification ====="

MAX_LOOP=5
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
	# Set the route defaults
	create_rule route-list $SCRIPTDIR/rules/default-route-rules.json $SCRIPTDIR/output/default-route-list.json
	create_rule route-list $SCRIPTDIR/rules/v2-selector-rule.json $SCRIPTDIR/output/route-list-v2.json

	export A8_GREMLIN_URL=http://localhost:31500

	$A8_CLI recipe-run --topology $SCRIPTDIR/../bookinfo-topology.json --scenarios $SCRIPTDIR/../bookinfo-gremlins.json --checks $SCRIPTDIR/../bookinfo-checks.json -r $SCRIPTDIR/gremlin-script/load-productpage.sh --header 'Cookie' --pattern='user=jason' -w 30s -f -o json > /tmp/recipe-run.json

	jsondiff_recipe_results $SCRIPTDIR/output/recipe-run-results.json /tmp/recipe-run.json
	if [ $? -gt 0 ]; then
                (( retry_count=retry_count+1 ))
		dump_debug
		echo "failed"
		echo "Actual a8ctl action-list output does not match the expected output."
                if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			cleanup_and_exit
		fi
		echo "Retrying...."
	else
		break
	fi
done

echo "Step 5 passed"

# Cleanup the rules
$A8_CLI rule-delete -a -f

if [ "$ENV" == "k8s" ]; then
	# k8s doesn't support the traffic commands
	exit 0
fi

echo "===== Step 6: Gradually migrate traffic to reviews:v3 for all users ====="

MAX_LOOP=5
retry_count=1
GATEWAY_URL=localhost:32000
COMMAND_INPUT="curl $GATEWAY_URL/reviews/health -s -I | grep \"Set-Cookie\" | sed 's/ /\n/g' | sed 's/\;//g' | grep version | cut -d '=' -f 2-"
EXPECTED_OUTPUT1="reviews-version=v1"
EXPECTED_OUTPUT2="reviews-version=v3"
while [  $retry_count -le $((MAX_LOOP)) ]; do
	create_rule route-list $SCRIPTDIR/rules/default-route-rules.json $SCRIPTDIR/output/default-route-list.json
	TRAFFIC_START_RESP=$($A8_CLI traffic-start -s reviews -v version=v3)
	if [ "$TRAFFIC_START_RESP" != "Transfer starting for \"reviews\": diverting 10% of traffic from \"version=v1\" to \"version=v3\"" ]; then
		echo "failed"
		echo "Failed to divert 10% traffic to v3"
		cleanup_and_exit
	fi

	# Give it a bit to process the request
	sleep 5

	# Validate that 10% of traffic is routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT1" "$EXPECTED_OUTPUT2" 10
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			cleanup_and_exit
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
	TRAFFIC_STEP_RESP=$($A8_CLI traffic-step -s reviews)
	if [ "$TRAFFIC_STEP_RESP" != "Transfer starting for \"reviews\": diverting 20% of traffic from \"version=v1\" to \"version=v3\"" ]; then
        	echo "failed"
        	echo "Failed to divert 20% of traffic to v3"
		cleanup_and_exit
	fi

	# Give it a bit to process the request
	sleep 5

	# Validate that 20% of traffic is routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT1" "$EXPECTED_OUTPUT2" 20
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			cleanup_and_exit
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
	TRAFFIC_STEP50_RESP=$($A8_CLI traffic-step -s reviews --amount 50)
	if [ "$TRAFFIC_STEP50_RESP" != "Transfer starting for \"reviews\": diverting 50% of traffic from \"version=v1\" to \"version=v3\"" ]; then
        	echo "failed"
        	echo "Failed to divert 50% of traffic to v3"
		cleanup_and_exit
	fi

	# Give it a bit to process the request
	sleep 5

	# Validate that 50% of traffic is routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT1" "$EXPECTED_OUTPUT2" 50
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			cleanup_and_exit
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
	TRAFFIC_STEP100_RESP=$($A8_CLI traffic-step -s reviews --amount 100)
	if [ "$TRAFFIC_STEP100_RESP" != "Transfer complete for \"reviews\": sending 100% of traffic to \"version=v3\"" ]; then
        	echo "failed"
        	echo "Failed to send 100% of traffic to v3"
		cleanup_and_exit
	fi

	# Give it a bit to process the request
	sleep 5

	# Validate that 100% of traffic is routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT1" "$EXPECTED_OUTPUT2" 100
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			cleanup_and_exit
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
	# The default route is now v3.  Start 50% traffic to v1.
	TRAFFIC_START_RESP=$($A8_CLI traffic-start -s reviews -v version=v1 -a 50)
	if [ "$TRAFFIC_START_RESP" != "Transfer starting for \"reviews\": diverting 50% of traffic from \"version=v3\" to \"version=v1\"" ]; then
		echo "failed"
		echo "Failed to divert 50% traffic to v1"
		cleanup_and_exit
	fi

	# Give it a bit to process the request
	sleep 5

	# Validate that 50% of traffic is routing to v1
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT2" "$EXPECTED_OUTPUT1" 50
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			cleanup_and_exit
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
while [  $retry_count -le $((MAX_LOOP)) ]; do
	TRAFFIC_ABORT_RESP=$($A8_CLI traffic-abort -s reviews)
	if [ "$TRAFFIC_ABORT_RESP" != "Transfer aborted for \"reviews\": all traffic reverted to \"version=v3\"" ]; then
        	echo "failed"
        	echo "Failed to abort"
		cleanup_and_exit
	fi

	# Give it a bit to process the request
	sleep 5

	# Validate that traffic is back routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT2" "$EXPECTED_OUTPUT1" 100
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			cleanup_and_exit
		fi
	else
		break
	fi
done

echo "Step 6 passed"

# Cleanup the rules
$A8_CLI rule-delete -a -f
