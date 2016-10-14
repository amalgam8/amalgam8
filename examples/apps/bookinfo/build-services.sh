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

set -o errexit

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

pushd $SCRIPTDIR/productpage
  docker build -t amalgam8/a8-examples-bookinfo-productpage:v1 .
  docker build -t amalgam8/a8-examples-bookinfo-productpage-sidecar:v1 -f Dockerfile.sidecar .
popd

pushd $SCRIPTDIR/details
  docker build -t amalgam8/a8-examples-bookinfo-details:v1 .
  docker build -t amalgam8/a8-examples-bookinfo-details-sidecar:v1 -f Dockerfile.sidecar .
popd

pushd $SCRIPTDIR/reviews
  #java build the app.
  docker run --rm -v `pwd`:/usr/bin/app:rw niaquinto/gradle clean build
  pushd reviews-wlpcfg
    #plain build -- no ratings
    docker build -t amalgam8/a8-examples-bookinfo-reviews:v1 --build-arg service_version=v1 .
    docker build -t amalgam8/a8-examples-bookinfo-reviews-sidecar:v1 --build-arg service_version=v1 -f Dockerfile.sidecar .
    #with ratings black stars
    docker build -t amalgam8/a8-examples-bookinfo-reviews:v2 --build-arg service_version=v2 --build-arg enable_ratings=true .
    docker build -t amalgam8/a8-examples-bookinfo-reviews-sidecar:v2 --build-arg service_version=v2 --build-arg enable_ratings=true -f Dockerfile.sidecar .
    #with ratings red stars
    docker build -t amalgam8/a8-examples-bookinfo-reviews:v3 --build-arg service_version=v3 --build-arg enable_ratings=true --build-arg star_color=red .
    docker build -t amalgam8/a8-examples-bookinfo-reviews-sidecar:v3 --build-arg service_version=v3 --build-arg enable_ratings=true --build-arg star_color=red -f Dockerfile.sidecar .
  popd
popd

pushd $SCRIPTDIR/ratings
  docker build -t amalgam8/a8-examples-bookinfo-ratings:v1 .
  docker build -t amalgam8/a8-examples-bookinfo-ratings-sidecar:v1 -f Dockerfile.sidecar .
popd
