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

A8_TEST_ENV=$2

# Remove the v from the version
APP_VER=$(echo $APP_VER | sed "s/v//")

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

GENERATED_DOCKERFILE_NAME=Dockerfile.generated.sidecar
A8_SIDECAR_SCRIPT_NAME=a8sidecar-testing.sh
A8_SIDECAR_SCRIPT=$SCRIPTDIR/../../../testing/$A8_SIDECAR_SCRIPT_NAME
A8_SIDECAR_TAR_NAME=a8sidecar-current.tar.gz
A8_SIDECAR_TAR=$SCRIPTDIR/../../../release/$A8_SIDECAR_TAR_NAME

generate_sidecar_dockerfile(){
  if [ "$A8_TEST_ENV" == "examples" ]; then
    cp $A8_SIDECAR_SCRIPT .
    cp $A8_SIDECAR_TAR .
    sed 's#{COPY_SCRIPT}#COPY '"$A8_SIDECAR_SCRIPT_NAME"' /opt/microservices/#; s#{COPY_SIDECAR_TAR}#COPY '"$A8_SIDECAR_TAR_NAME"' /opt/microservices/#; s#RUN wget -qO- https://github.com/amalgam8/amalgam8/releases/download/{A8_RELEASE}/a8sidecar.sh | sh#'"RUN /opt/microservices/$A8_SIDECAR_SCRIPT_NAME"'#' Dockerfile.sidecar > $GENERATED_DOCKERFILE_NAME
  else
    sed 's/{COPY_SCRIPT}//; s/{COPY_SIDECAR_TAR}//; s/{A8_RELEASE}/'"v$APP_VER"'/' Dockerfile.sidecar > $GENERATED_DOCKERFILE_NAME
  fi
}

cleanup_generated_files(){
  if [ "$A8_TEST_ENV" == "examples" ]; then
    rm $GENERATED_DOCKERFILE_NAME $A8_SIDECAR_SCRIPT_NAME $A8_SIDECAR_TAR_NAME
  else
    rm $GENERATED_DOCKERFILE_NAME
  fi
}

echo "Building productpage"
pushd $SCRIPTDIR/productpage
  docker build -t amalgam8/a8-examples-bookinfo-productpage-v1:$APP_VER .
  generate_sidecar_dockerfile
  docker build -t amalgam8/a8-examples-bookinfo-productpage-sidecar-v1:$APP_VER -f $GENERATED_DOCKERFILE_NAME .
  cleanup_generated_files
popd

echo "Building details"
pushd $SCRIPTDIR/details
  docker build -t amalgam8/a8-examples-bookinfo-details-v1:$APP_VER .
  generate_sidecar_dockerfile
  docker build -t amalgam8/a8-examples-bookinfo-details-sidecar-v1:$APP_VER -f $GENERATED_DOCKERFILE_NAME .
  cleanup_generated_files
popd

pushd $SCRIPTDIR/reviews
  #java build the app.
  docker run --rm -v `pwd`:/usr/bin/app:rw niaquinto/gradle clean build
  pushd reviews-wlpcfg
    #plain build -- no ratings
    docker build -t amalgam8/a8-examples-bookinfo-reviews-v1:$APP_VER --build-arg service_version=v1 .
    generate_sidecar_dockerfile
    docker build -t amalgam8/a8-examples-bookinfo-reviews-sidecar-v1:$APP_VER --build-arg service_version=v1 -f $GENERATED_DOCKERFILE_NAME .
    #with ratings black stars
    docker build -t amalgam8/a8-examples-bookinfo-reviews-v2:$APP_VER --build-arg service_version=v2 --build-arg enable_ratings=true .
    docker build -t amalgam8/a8-examples-bookinfo-reviews-sidecar-v2:$APP_VER --build-arg service_version=v2 --build-arg enable_ratings=true -f $GENERATED_DOCKERFILE_NAME .
    #with ratings red stars
    docker build -t amalgam8/a8-examples-bookinfo-reviews-v3:$APP_VER --build-arg service_version=v3 --build-arg enable_ratings=true --build-arg star_color=red .
    docker build -t amalgam8/a8-examples-bookinfo-reviews-sidecar-v3:$APP_VER --build-arg service_version=v3 --build-arg enable_ratings=true --build-arg star_color=red -f $GENERATED_DOCKERFILE_NAME .
    cleanup_generated_files
  popd
popd

pushd $SCRIPTDIR/ratings
  docker build -t amalgam8/a8-examples-bookinfo-ratings-v1:$APP_VER .
  generate_sidecar_dockerfile
  docker build -t amalgam8/a8-examples-bookinfo-ratings-sidecar-v1:$APP_VER -f $GENERATED_DOCKERFILE_NAME .
  cleanup_generated_files
popd
