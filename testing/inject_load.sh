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

echoerr() {
    echo "$@" 1>&2;
}
echostart() {
    echo -n "$@" 1>&2;
}
echoend() {
    echoerr $@;
}
sleep 10
echostart "injecting traffic for user=shriram, expecting productpage_v1 in less than 2s.."
before=$(date +"%s")
curl -s -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/productpage_no_rulematch.html
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 2 ]; then
    echoerr "failed"
    echoerr "Productpage took more $delta seconds to respond (expected 2s) for user=shriram in gremlin test phase"
    exit 1
fi

diff -u $SCRIPTDIR/productpage_v1.html /tmp/productpage_no_rulematch.html 1>&2
if [ $? -gt 0 ]; then
    echoerr "failed"
    echoerr "Productpage does not match productpage_v1 after injecting fault rule for user=shriram in gremlin test phase"
    exit 1
fi
echoend "works!"

####For jason
echostart "injecting traffic for user=jason, expecting 'reviews unavailable' message in approx 7s.."
before=$(date +"%s")
curl -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage >/tmp/productpage_rulematch.html
after=$(date +"%s")

delta=$(($after-$before))
if [ $delta -gt 8 -o $delta -lt 5 ]; then
    echoerr "failed"
    echoerr "Productpage took more $delta seconds to respond (expected <8s && >5s) for user=jason in gremlin test phase"
    exit 1
fi

diff -u $SCRIPTDIR/productpage_rulematch.html /tmp/productpage_rulematch.html 1>&2
if [ $? -gt 0 ]; then
    echoerr "failed"
    echoerr "Productpage does not match productpage_rulematch.html after injecting fault rule for user=jason in gremlin test phase"
    exit 1
fi
echoend "works!"
sleep 15
