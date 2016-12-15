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

jsondiff() {
    # Sort JSON fields, or fallback to original text
    file1=$(jq -S . $1 2> /dev/null) || file1=$(cat $1)
    file2=$(jq -S . $2 2> /dev/null) || file2=$(cat $2)
    
    # Diff, but ignore all whitespace since
    diff -Ewb <(echo $file1) <(echo $file2)        
}

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

a8ctl rule-clear
sleep 2

######Version Routing##############
echo "testing version routing.."
a8ctl route-set --default v1 productpage
a8ctl route-set --default v1 details
a8ctl route-set --default v1 ratings
a8ctl route-set --default v1 --selector 'v2(user="jason")' reviews
sleep 10

echo -n "injecting traffic for user=shriram, expecting productpage_v1.."
curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/productpage_v1.json

jsondiff $SCRIPTDIR/productpage_v1.json /tmp/productpage_v1.json
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage not match productpage_v1 after setting route to reviews:v2 for user=shriram in version routing test phase"
    exit 1
fi
echo "works!"

echo -n "injecting traffic for user=jason, expecting productpage_v2.."
curl -s -b 'foo=bar;user=jason;ding=dong;x' http://localhost:32000/productpage/productpage >/tmp/productpage_v2.json

jsondiff $SCRIPTDIR/productpage_v2.json /tmp/productpage_v2.json
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_v2 after setting route to reviews:v2 for user=jason in version routing test phase"
    exit 1
fi
echo "works!"

########Fault injection
echo "testing fault injection.."
a8ctl rule-set --source reviews:v2 --destination ratings:v1 --header Cookie --pattern 'user=jason' --delay-probability 1.0 --delay 7
sleep 10

###For shriram
echo -n "injecting traffic for user=shriram, expecting productpage_v1 in less than 2s.."
before=$(date +"%s")
curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/productpage_no_rulematch.json
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 2 ]; then
    echo "failed"
    echo "Productpage took more $delta seconds to respond (expected 2s) for user=shriram in fault injection phase"
    exit 1
fi

jsondiff $SCRIPTDIR/productpage_v1.json /tmp/productpage_no_rulematch.json
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
a8ctl rule-clear
sleep 10

echo -n "Testing app again for user=jason, expecting productpage_v2 in less than 2s.."
before=$(date +"%s")
curl -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage >/tmp/productpage_v2.json
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 2 ]; then
    echo "failed"
    echo "Productpage took more $delta seconds to respond (expected <2s) after clearing rules for user=jason in fault injection phase"
    exit 1
fi

jsondiff $SCRIPTDIR/productpage_v2.json /tmp/productpage_v2.json
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_v2.json after clearing fault rules for user=jason in fault injection test phase"
    exit 1
fi
echo "works!"
