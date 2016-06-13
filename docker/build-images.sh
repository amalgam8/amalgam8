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



# Amalgam8 Public Images

# a8-registry-0.1                               - Amalgam8 registry
# a8-controller-0.1                             - Amalgam8 controller
# a8-sidecar-reg-0.1                            - Amalgam8 service registration-only sidecar image
# a8-sidecar-proxy-0.1                          - Amalgam8 service registration and proxy sidecar image
# a8-appbase-reg-0.1                            - Amalgam8 service registration-only application base image
# a8-appbase-proxy-0.1                          - Amalgam8 service registration and proxy application base image

# a8-examples-helloworld
# a8-examples-bookinfo-productpage:v1
# a8-examples-bookinfo-details:v1
# a8-examples-bookinfo-reviews:v1
# a8-examples-bookinfo-reviews:v2
# a8-examples-bookinfo-reviews:v3
# a8-examples-bookinfo-ratings:v1

# a8-examples-helloworld-sidecar
# a8-examples-bookinfo-productpage-sidecar:v1
# a8-examples-bookinfo-details-sidecar:v1
# a8-examples-bookinfo-reviews-sidecar:v1
# a8-examples-bookinfo-reviews-sidecar:v2
# a8-examples-bookinfo-reviews-sidecar:v3
# a8-examples-bookinfo-ratings-sidecar:v1

set -o errexit

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

pushd $SCRIPTDIR/../apps/helloworld
  docker build -t a8-examples-helloworld .
  docker build -t a8-examples-helloworld-sidecar -f Dockerfile.sidecar .
popd

pushd $SCRIPTDIR/../apps/bookinfo/productpage
  docker build -t a8-examples-bookinfo-productpage:v1 .
  docker build -t a8-examples-bookinfo-productpage-sidecar:v1 -f Dockerfile.sidecar .
popd

pushd $SCRIPTDIR/../apps/bookinfo/details
  docker build -t a8-examples-bookinfo-details:v1 .
  docker build -t a8-examples-bookinfo-details-sidecar:v1 -f Dockerfile.sidecar .
popd

pushd $SCRIPTDIR/../apps/bookinfo/reviews
  #plain build -- no ratings
  docker build -t examples-bookinfo-reviews:v1 .
  docker build -t examples-bookinfo-reviews-sidecar:v1 -f Dockerfile.sidecar.

  #with ratings black stars
  docker build -t examples-bookinfo-reviews:v2 --build-arg enable_ratings=true .
  docker build -t examples-bookinfo-reviews-sidecar:v2 --build-arg enable_ratings=true -f Dockerfile.sidecar .

  #with ratings red stars
  docker build -t examples-bookinfo-reviews:v3 --build-arg enable_ratings=true --build-arg star_color=red .
  docker build -t examples-bookinfo-reviews-sidecar:v3 --build-arg enable_ratings=true --build-arg star_color=red -f Dockerfile.sidecar .
popd

pushd $SCRIPTDIR/../apps/bookinfo/ratings
  docker build -t a8-examples-bookinfo-ratings:v1 .
  docker build -t a8-examples-bookinfo-ratings-sidecar:v1 -f Dockerfile.sidecar .
popd
