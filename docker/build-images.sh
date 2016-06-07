#!/bin/bash

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
