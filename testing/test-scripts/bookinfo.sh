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

set -x
set -o errexit

# Set env vars
export A8_CONTROLLER_URL=http://localhost:31200
export A8_REGISTRY_URL=http://localhost:31300
export A8_GATEWAY_URL=http://localhost:32000
export A8_LOG_SERVER=http://localhost:30200
export A8_GREMLIN_URL=http://localhost:31500

# Set local vars
A8_TEST_SUITE=$1
if [ "$A8_TEST_SUITE" == "examples" ]; then
    PRODUCTPAGE_V1_OUTPUT="productpage_v1.html"
    PRODUCTPAGE_V2_OUTPUT="productpage_v2.html"
else
    PRODUCTPAGE_V1_OUTPUT="productpage_v1.json"
    PRODUCTPAGE_V2_OUTPUT="productpage_v2.json"
fi
# docker or k8s
if [ -z "$A8_CONTAINER_ENV" ]; then
    A8_CONTAINER_ENV="docker"
fi

if [ -z "$A8_TEST_GREMLIN" ]; then
    A8_TEST_GREMLIN="true"
fi

jsondiff() {
    # Sort JSON fields, or fallback to original text
    file1=$(jq -S . $1 2> /dev/null) || file1=$(cat $1)
    file2=$(jq -S . $2 2> /dev/null) || file2=$(cat $2)

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
			#dump_debug
			return 1
		elif [ $retry_count -eq 100 ]; then
			echo "  Passed test"
			echo ""
		fi
	done
	return 0
}

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
CLIBIN=$SCRIPTDIR/../../bin/a8ctl-beta-linux

$CLIBIN rule-delete -a -f
sleep 2

######Version Routing##############
echo "testing version routing.."
$CLIBIN rule-create -f $SCRIPTDIR/default_routes.yaml
$CLIBIN rule-create -f $SCRIPTDIR/route_jason_v2.yaml
sleep 10

echo -n "injecting traffic for user=shriram, expecting productpage_v1.."
curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/$PRODUCTPAGE_V1_OUTPUT

jsondiff $SCRIPTDIR/$PRODUCTPAGE_V1_OUTPUT /tmp/$PRODUCTPAGE_V1_OUTPUT
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage not match productpage_v1 after setting route to reviews:v2 for user=shriram in version routing test phase"
    exit 1
fi
echo "works!"

echo -n "injecting traffic for user=jason, expecting productpage_v2.."
curl -s -b 'foo=bar;user=jason;ding=dong;x' http://localhost:32000/productpage/productpage >/tmp/$PRODUCTPAGE_V2_OUTPUT

jsondiff $SCRIPTDIR/$PRODUCTPAGE_V2_OUTPUT /tmp/$PRODUCTPAGE_V2_OUTPUT
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_v2 after setting route to reviews:v2 for user=jason in version routing test phase"
    exit 1
fi
echo "works!"

########Fault injection
echo "testing fault injection.."
output=$( $CLIBIN rule-create -f $SCRIPTDIR/fault_injection.yaml )

fault_rule_id=$( tr -d ' \t\n\r' <<< "$output" | sed -e 's/\({"ids":\["\)//g' -e 's/\("\]}\)//g' )
sleep 10

###For shriram
echo -n "injecting traffic for user=shriram, expecting productpage_v1 in less than 2s.."
before=$(date +"%s")
curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/$PRODUCTPAGE_V1_OUTPUT
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 2 ]; then
    echo "failed"
    echo "Productpage took more $delta seconds to respond (expected 2s) for user=shriram in fault injection phase"
    exit 1
fi

jsondiff $SCRIPTDIR/$PRODUCTPAGE_V1_OUTPUT /tmp/$PRODUCTPAGE_V1_OUTPUT
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_v1 after injecting fault rule for user=shriram in fault injection test phase"
    exit 1
fi
echo "works!"

####For jason
echo -n "injecting traffic for user=jason, expecting 'reviews unavailable' message in approx 7s.."
before=$(date +"%s")
curl -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage >/tmp/productpage_rulematch.json
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 8 -o $delta -lt 5 ]; then
    echo "failed"
    echo "Productpage took more $delta seconds to respond (expected <8s && >5s) for user=jason in fault injection phase"
    exit 1
fi

jsondiff $SCRIPTDIR/productpage_rulematch.json /tmp/productpage_rulematch.json
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_rulematch.json after injecting fault rule for user=jason in fault injection test phase"
    exit 1
fi
echo "works!"

#####Clear rules and check
echo "clearing all rules.."
$CLIBIN rule-delete -i $fault_rule_id
sleep 15

echo -n "Testing app again for user=jason, expecting productpage_v2 in less than 2s.."
before=$(date +"%s")
curl -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage >/tmp/$PRODUCTPAGE_V2_OUTPUT
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 2 ]; then
    echo "failed"
    echo "Productpage took more $delta seconds to respond (expected <2s) after clearing rules for user=jason in fault injection phase"
    exit 1
fi

jsondiff $SCRIPTDIR/$PRODUCTPAGE_V2_OUTPUT /tmp/$PRODUCTPAGE_V2_OUTPUT
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_v2.json after clearing fault rules for user=jason in fault injection test phase"
    exit 1
fi
echo "works!"

#######Gremlin
if [ "$A8_TEST_GREMLIN" = true ]; then
  GREMLIN_FILES=$SCRIPTDIR/../../examples
  echo "Testing gremlin recipe.."
  echo "please wait.."
  MAX_LOOP=5
  retry_count=1
  while [  $retry_count -le $((MAX_LOOP)) ]; do
  	$CLIBIN recipe-run -t $GREMLIN_FILES/bookinfo-topology.json -s $GREMLIN_FILES/bookinfo-gremlins.json -c $GREMLIN_FILES/bookinfo-checks.json -r $SCRIPTDIR/load-productpage.sh --header 'Cookie' --pattern='^(.*?;)?(user=jason)(;.*)?$' -w 30s -f -o json | jq '.' > /tmp/recipe-run-results.json

  	jsondiff $SCRIPTDIR/recipe-run-results.json /tmp/recipe-run-results.json
  	if [ $? -gt 0 ]; then
  		(( retry_count=retry_count+1 ))
  		echo "failed"
  		echo "Recipe-run-results.json does not match the expected output."
  		echo "Retrying.."
  	else
  		break
  	fi
  done
  echo "works!"
else
  echo "Skip Gremlin recipe test.."
fi

if [ "$A8_CONTAINER_ENV" == "k8s" ]; then
	# k8s doesn't support the traffic commands
	exit 0
fi

#######Gradually migrate traffic to reviews:v3 for all users
$CLIBIN rule-delete -a -f

# Quiet things down a bit
set +x

$CLIBIN rule-create -f $SCRIPTDIR/default_routes.yaml
sleep 10

MAX_LOOP=5
retry_count=1
SLEEP_TIME=3
GATEWAY_URL=localhost:32000
COMMAND_INPUT="curl $GATEWAY_URL/reviews/health -s -I | grep \"Set-Cookie\" | sed 's/ /\n/g' | sed 's/\;//g' | grep version | cut -d '=' -f 2-"
EXPECTED_OUTPUT1="reviews-version=v1"
EXPECTED_OUTPUT2="reviews-version=v3"
echo "Setting traffic-start.  Expecting 10% of traffic to be diverted to v3."
while [  $retry_count -le $((MAX_LOOP)) ]; do
	TRAFFIC_START_RESP=$($CLIBIN traffic-start -s reviews -v version=v3)
	if [ "$TRAFFIC_START_RESP" != "Transfer starting for \"reviews\": diverting 10% of traffic from \"version=v1\" to \"version=v3\"" ]; then
		echo "failed"
		echo "Failed to divert 10% traffic to v3"
		exit 1
	fi

	# Give it a bit to process the request
	sleep $SLEEP_TIME

	# Validate that 10% of traffic is routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT1" "$EXPECTED_OUTPUT2" 10
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			exit 1
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
echo "Setting traffic-step.  Expecting 20% of traffic to be diverted to v3."
while [  $retry_count -le $((MAX_LOOP)) ]; do
	TRAFFIC_STEP_RESP=$($CLIBIN traffic-step -s reviews)
	if [ "$TRAFFIC_STEP_RESP" != "Transfer starting for \"reviews\": diverting 20% of traffic from \"version=v1\" to \"version=v3\"" ]; then
		echo "failed"
		echo "Failed to divert 20% of traffic to v3"
		exit 1
	fi

	# Give it a bit to process the request
	sleep $SLEEP_TIME

	# Validate that 20% of traffic is routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT1" "$EXPECTED_OUTPUT2" 20
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			exit 1
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
echo "Setting traffic-step.  Expecting 50% of traffic to be diverted to v3."
while [  $retry_count -le $((MAX_LOOP)) ]; do
	TRAFFIC_STEP50_RESP=$($CLIBIN traffic-step -s reviews --amount 50)
	if [ "$TRAFFIC_STEP50_RESP" != "Transfer starting for \"reviews\": diverting 50% of traffic from \"version=v1\" to \"version=v3\"" ]; then
		echo "failed"
		echo "Failed to divert 50% of traffic to v3"
		exit 1
	fi

	# Give it a bit to process the request
	sleep $SLEEP_TIME

	# Validate that 50% of traffic is routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT1" "$EXPECTED_OUTPUT2" 50
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			exit 1
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
echo "Setting traffic-step.  Expecting 100% of traffic to be diverted to v3."
while [  $retry_count -le $((MAX_LOOP)) ]; do
	TRAFFIC_STEP100_RESP=$($CLIBIN traffic-step -s reviews --amount 100)
	if [ "$TRAFFIC_STEP100_RESP" != "Transfer complete for \"reviews\": sending 100% of traffic to \"version=v3\"" ]; then
		echo "failed"
		echo "Failed to send 100% of traffic to v3"
		exit 1
	fi

	# Give it a bit to process the request
	sleep $SLEEP_TIME

	# Validate that 100% of traffic is routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT1" "$EXPECTED_OUTPUT2" 100
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			exit 1
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
echo "Setting traffic-start.  Expecting 50% of traffic to be diverted to v1."
while [  $retry_count -le $((MAX_LOOP)) ]; do
	# The default route is now v3.  Start 50% traffic to v1.
	TRAFFIC_START_RESP=$($CLIBIN traffic-start -s reviews -v version=v1 -a 50)
	if [ "$TRAFFIC_START_RESP" != "Transfer starting for \"reviews\": diverting 50% of traffic from \"version=v3\" to \"version=v1\"" ]; then
		echo "failed"
		echo "Failed to divert 50% traffic to v1"
		exit 1
	fi

	# Give it a bit to process the request
	sleep $SLEEP_TIME

	# Validate that 50% of traffic is routing to v1
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT2" "$EXPECTED_OUTPUT1" 50
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			exit 1
		fi
	else
		break
	fi
done

MAX_LOOP=5
retry_count=1
echo "Setting traffic-abort.  Expecting 100% of traffic to be reverted to v3."
while [  $retry_count -le $((MAX_LOOP)) ]; do
	TRAFFIC_ABORT_RESP=$($CLIBIN traffic-abort -s reviews)
	if [ "$TRAFFIC_ABORT_RESP" != "Transfer aborted for \"reviews\": all traffic reverted to \"version=v3\"" ]; then
		echo "failed"
		echo "Failed to abort"
		exit 1
	fi

	# Give it a bit to process the request
	sleep $SLEEP_TIME

	# Validate that traffic is back routing to v3
	# Curl the health check and check the version cookie
	check_routing_rules "$COMMAND_INPUT" "$EXPECTED_OUTPUT2" "$EXPECTED_OUTPUT1" 100
	if [ $? -gt 0 ]; then
		(( retry_count=retry_count+1 ))
		if [ $retry_count -eq $((MAX_LOOP+1)) ]; then
			exit 1
		fi
	else
		break
	fi
done

echo "done"
