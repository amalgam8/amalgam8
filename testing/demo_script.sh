#!/bin/bash

######Version Routing##############
echo "testing version routing.."
a8ctl route-set --default v1 productpage
a8ctl route-set --default v1 details
a8ctl route-set --default v1 ratings
a8ctl route-set --default v1 --selector 'v2(user="jason")' reviews
sleep 2

echo -n "injecting traffic for user=shriram, expecting productpage_v1.."
curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/productpage_v1.html

diff -u productpage_v1.html /tmp/productpage_v1.html
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage not match productpage_v1 after setting route to reviews:v2 for user=shriram in version routing test phase"
    exit 1
fi
echo "works!"

echo -n "injecting traffic for user=jason, expecting productpage_v2.."
curl -s -b 'foo=bar;user=jason;ding=dong;x' http://localhost:32000/productpage/productpage >/tmp/productpage_v2.html

diff -u productpage_v2.html /tmp/productpage_v2.html
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_v2 after setting route to reviews:v2 for user=jason in version routing test phase"
    exit 1
fi
echo "works!"

########Fault injection
echo "testing fault injection.."
a8ctl rule-set --source reviews:v2 --destination ratings:v1 --header Cookie --pattern 'user=jason' --delay-probability 1.0 --delay 7
sleep 2

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

diff -u productpage_v1.html /tmp/productpage_no_rulematch.html
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

diff -u productpage_rulematch.html /tmp/productpage_rulematch.html
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_rulematch.html after injecting fault rule for user=jason in fault injection test phase"
    exit 1
fi
echo "works!"

#####Clear rules and check
echo "clearing all rules.."
a8ctl rule-clear
sleep 2

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

diff -u productpage_v2.html /tmp/productpage_v2.html
if [ $? -gt 0 ]; then
    echo "failed"
    echo "Productpage does not match productpage_v2.html after clearing fault rules for user=jason in fault injection test phase"
    exit 1
fi
echo "works!"

#######Gremlin
echo "Testing gremlin recipe.."
a8ctl recipe-run --topology ../apps/bookinfo/topology.json --scenarios ../apps/bookinfo/gremlins.json --checks ../apps/bookinfo/checklist.json --run-load-script ./inject_load.sh --header 'Cookie' --pattern='user=jason'
echo "Gremlin tests successful.."
