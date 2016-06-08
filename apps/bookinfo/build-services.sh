#!/bin/bash

#gunicorn -D -w 1 -b 0.0.0.0:10081 --reload details:app
#gunicorn -D -w 1 -b 0.0.0.0:10082 --reload reviews:app
#gunicorn -w 1 -b 0.0.0.0:19080 --reload --access-logfile prod.log --error-logfile prod.log productpage:app >>prod.log 2>&1 &

set -o errexit

pushd productpage
  if [ -n "$GATEWAY_URL" ]
  then
    sed -e "s#http://192.168.33.33:32000#${GATEWAY_URL}#" Dockerfile > Dockerfile.tmp
  else
    cp Dockerfile Dockerfile.tmp
  fi
  docker build -t a8-examples-bookinfo-productpage:v1 -f Dockerfile.tmp .
popd

pushd details
  docker build -t a8-examples-bookinfo-details:v1 .
popd

pushd reviews
  #plain build -- no ratings
  docker build -t a8-examples-bookinfo-reviews:v1 .
  #with ratings black stars
  docker build -t a8-examples-bookinfo-reviews:v2 --build-arg enable_ratings=true .
  #with ratings red stars
  docker build -t a8-examples-bookinfo-reviews:v3 --build-arg enable_ratings=true --build-arg star_color=red .
popd

pushd ratings
  docker build -t a8-examples-bookinfo-ratings:v1 .
popd
