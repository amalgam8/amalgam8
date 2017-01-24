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

APP_VER=$1
if [ -z $APP_VER ]; then
  echo "The version must be set."
  exit 1
fi

if [ "$APP_VER" == "unknown" ]; then
  echo "The version cannot be unknown."
  exit 1
fi

A8_SIDECAR_RELEASE=$2
if [ -z $A8_SIDECAR_RELEASE ]; then
  echo "The sidecar release cannot be blank."
  exit 1
fi

# Remove the v from the version
APP_VER=$(echo $APP_VER | sed "s/v//")

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

A8_SIDECAR_TAR_NAME=${A8_SIDECAR_RELEASE}.tar.gz
A8_SIDECAR_TAR=$SCRIPTDIR/../../../release/$A8_SIDECAR_TAR_NAME

copy_sidecar_tar(){
  cp $A8_SIDECAR_TAR .
}

cleanup_sidecar_tar(){
  rm $A8_SIDECAR_TAR_NAME
}

echo "Building productpage"
pushd $SCRIPTDIR/productpage
  docker build -t amalgam8/a8-examples-bookinfo-productpage-v1:$APP_VER .
  copy_sidecar_tar
  docker build -t amalgam8/a8-examples-bookinfo-productpage-sidecar-v1:$APP_VER -f Dockerfile.sidecar --build-arg A8_SIDECAR_RELEASE=$A8_SIDECAR_RELEASE .
  cleanup_sidecar_tar
popd

echo "Building details"
pushd $SCRIPTDIR/details
  docker build -t amalgam8/a8-examples-bookinfo-details-v1:$APP_VER .
  copy_sidecar_tar
  docker build -t amalgam8/a8-examples-bookinfo-details-sidecar-v1:$APP_VER -f Dockerfile.sidecar --build-arg A8_SIDECAR_RELEASE=$A8_SIDECAR_RELEASE .
  cleanup_sidecar_tar
popd

pushd $SCRIPTDIR/reviews
  #java build the app.
  docker run --rm -v `pwd`:/usr/bin/app:rw niaquinto/gradle clean build
  pushd reviews-wlpcfg
    copy_sidecar_tar
    #plain build -- no ratings
    docker build -t amalgam8/a8-examples-bookinfo-reviews-v1:$APP_VER --build-arg service_version=v1 .
    docker build -t amalgam8/a8-examples-bookinfo-reviews-sidecar-v1:$APP_VER --build-arg service_version=v1 --build-arg A8_SIDECAR_RELEASE=$A8_SIDECAR_RELEASE -f Dockerfile.sidecar .
    #with ratings black stars
    docker build -t amalgam8/a8-examples-bookinfo-reviews-v2:$APP_VER --build-arg service_version=v2 --build-arg enable_ratings=true .
    docker build -t amalgam8/a8-examples-bookinfo-reviews-sidecar-v2:$APP_VER --build-arg service_version=v2 --build-arg enable_ratings=true --build-arg A8_SIDECAR_RELEASE=$A8_SIDECAR_RELEASE -f Dockerfile.sidecar .
    #with ratings red stars
    docker build -t amalgam8/a8-examples-bookinfo-reviews-v3:$APP_VER --build-arg service_version=v3 --build-arg enable_ratings=true --build-arg star_color=red .
    docker build -t amalgam8/a8-examples-bookinfo-reviews-sidecar-v3:$APP_VER --build-arg service_version=v3 --build-arg enable_ratings=true --build-arg star_color=red --build-arg A8_SIDECAR_RELEASE=$A8_SIDECAR_RELEASE -f Dockerfile.sidecar .
    cleanup_sidecar_tar
  popd
popd

pushd $SCRIPTDIR/ratings
  copy_sidecar_tar
  docker build -t amalgam8/a8-examples-bookinfo-ratings-v1:$APP_VER .
  docker build -t amalgam8/a8-examples-bookinfo-ratings-sidecar-v1:$APP_VER --build-arg A8_SIDECAR_RELEASE=$A8_SIDECAR_RELEASE -f Dockerfile.sidecar .
  cleanup_sidecar_tar
popd
