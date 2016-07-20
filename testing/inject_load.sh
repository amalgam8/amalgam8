#!/bin/bash

curl -b 'foo=bar;user=shriram;x' http://localhost:32000/productpage/productpage >/tmp/productpage_no_rulematch.html
diff productpage_v1.html /tmp/productpage_no_rulematch.html
if [ $? -gt 0 ]; then
    echo "Productpage does not match productpage_v1 after injecting fault rule for user=shriram in gremlin test phase"
    exit 1
fi
curl -b 'foo=bar;user=jason;x' http://localhost:32000/productpage/productpage >/tmp/productpage_rulematch.html
diff productpage_rulematch.html /tmp/productpage_rulematch.html
if [ $? -gt 0 ]; then
    echo "Productpage does not match productpage_rulematch.html after injecting fault rule for user=jason in gremlin test phase"
    exit 1
fi
sleep 5
