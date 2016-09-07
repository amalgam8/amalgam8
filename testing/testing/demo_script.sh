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

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
RECIPE_PATH=$GOPATH/src/github.com/amalgam8/amalgam8/examples/apps/bookinfo
a8ctl service-list
sleep 2
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
curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/productpage_v1.html

diff -u $SCRIPTDIR/productpage_v1.html /tmp/productpage_v1.html
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage not match productpage_v1 after setting route to reviews:v2 for user=shriram in version routing test phase"
    exit 1
fi
echo "works!"

echo -n "injecting traffic for user=jason, expecting productpage_v2.."
curl -s -b 'foo=bar;user=jason;ding=dong;x' http://localhost:32000/productpage/productpage >/tmp/productpage_v2.html

diff -u $SCRIPTDIR/productpage_v2.html /tmp/productpage_v2.html
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
curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/productpage_no_rulematch.html
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 2 ]; then
    echo "failed"
    echo "Productpage took more $delta seconds to respond (expected 2s) for user=shriram in fault injection phase"
    exit 1
fi

diff -u $SCRIPTDIR/productpage_v1.html /tmp/productpage_no_rulematch.html
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_v1 after injecting fault rule for user=shriram in fault injection test phase"
    exit 1
fi
echo "works!"

####For jason
echo -n "injecting traffic for user=jason, expecting 'reviews unavailable' message in approx 7s.."
before=$(date +"%s")
curl -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage >/tmp/productpage_rulematch.html
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 8 -o $delta -lt 5 ]; then
    echo "failed"
    echo "Productpage took more $delta seconds to respond (expected <8s && >5s) for user=jason in fault injection phase"
    exit 1
fi

diff -u $SCRIPTDIR/productpage_rulematch.html /tmp/productpage_rulematch.html
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_rulematch.html after injecting fault rule for user=jason in fault injection test phase"
    exit 1
fi
echo "works!"

#####Clear rules and check
echo "clearing all rules.."
a8ctl rule-clear
sleep 10

echo -n "Testing app again for user=jason, expecting productpage_v2 in less than 2s.."
before=$(date +"%s")
curl -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage >/tmp/productpage_v2.html
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 2 ]; then
    echo "failed"
    echo "Productpage took more $delta seconds to respond (expected <2s) after clearing rules for user=jason in fault injection phase"
    exit 1
fi

diff -u $SCRIPTDIR/productpage_v2.html /tmp/productpage_v2.html
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_v2.html after clearing fault rules for user=jason in fault injection test phase"
    exit 1
fi
echo "works!"

#######Gremlin
echo "Testing gremlin recipe.."
a8ctl recipe-run --topology $RECIPE_PATH/topology.json --scenarios $RECIPE_PATH/gremlins.json --checks $RECIPE_PATH/checklist.json --run-load-script $SCRIPTDIR/inject_load.sh --header 'Cookie' --pattern='user=jason' > /tmp/gremlin_results.txt
if [ $? -gt 0 ]; then
    echo "a8ctl recipe-run exited with non-zero status. Either load injection failed or there was an error in a8ctl"
    exit 1
fi

passes=$(grep "PASS" /tmp/gremlin_results.txt | wc -l|tr -d ' ')
if [ "$passes" != "3" ]; then
    echo "Gremlin tests failed: result does not have the expected number of passing assertions"
    cat /tmp/gremlin_results.txt
    exit 1
fi
failures=$(grep "FAIL" /tmp/gremlin_results.txt | wc -l|tr -d ' ')
if [ "$failures" != "1" ]; then
    echo "Gremlin tests failed: result does not have the expected number of failing assertions"
    cat /tmp/gremlin_results.txt
    exit 1
fi
echo "Gremlin tests were successful.."
