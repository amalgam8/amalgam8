#!/bin/bash
echoerr() {
    echo "$@" 1>&2;
}
echostart() {
    echo -n "$@" 1>&2;
}
echoend() {
    echoerr $@;
}

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

diff -u productpage_v1.html /tmp/productpage_no_rulematch.html 1>&2
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

diff -u productpage_rulematch.html /tmp/productpage_rulematch.html 1>&2
if [ $? -gt 0 ]; then
    echoerr "failed"
    echoerr "Productpage does not match productpage_rulematch.html after injecting fault rule for user=jason in gremlin test phase"
    exit 1
fi
echoend "works!"
sleep 5
