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
