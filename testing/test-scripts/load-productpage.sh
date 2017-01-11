#!/bin/bash

# This script runs after the a8ctl recipe-run sets the rules, but before the logs
# are checked.
sleep 20
curl -s -b 'foo=bar;user=jason;x' http://localhost:32000/productpage > /dev/null
