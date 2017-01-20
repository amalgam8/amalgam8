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


set -x

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
    cp $A8_SIDECAR_SCRIPT $SCRIPTDIR/
    cp $A8_SIDECAR_TAR $SCRIPTDIR/
    sed 's#{COPY_SCRIPT}#COPY '"$A8_SIDECAR_SCRIPT_NAME"' /opt/microservices#; s#{COPY_SIDECAR_TAR}#COPY '"$A8_SIDECAR_TAR_NAME"' /opt/microservices/#; s#RUN wget -qO- https://github.com/amalgam8/amalgam8/releases/download/{A8_RELEASE}/a8sidecar.sh | sh#'"RUN /opt/microservices/$A8_SIDECAR_SCRIPT_NAME"'#' $SCRIPTDIR/Dockerfile.sidecar > $SCRIPTDIR/$GENERATED_DOCKERFILE_NAME
  else
    sed 's/{COPY_SCRIPT}//; s/{COPY_SIDECAR_TAR}//; s/{A8_RELEASE}/'"v$APP_VER"'/' $SCRIPTDIR/Dockerfile.sidecar > $SCRIPTDIR/$GENERATED_DOCKERFILE_NAME
  fi
}

#################################################################################
# Build the helloworld image
#################################################################################
docker build -t amalgam8/a8-examples-helloworld-v1:$APP_VER $SCRIPTDIR
docker build -t amalgam8/a8-examples-helloworld-v2:$APP_VER $SCRIPTDIR

generate_sidecar_dockerfile
docker build -t amalgam8/a8-examples-helloworld-sidecar-v1:$APP_VER -f $SCRIPTDIR/$GENERATED_DOCKERFILE_NAME $SCRIPTDIR
docker build -t amalgam8/a8-examples-helloworld-sidecar-v2:$APP_VER -f $SCRIPTDIR/$GENERATED_DOCKERFILE_NAME $SCRIPTDIR
if [ "$A8_TEST_ENV" == "examples" ]; then
  rm $SCRIPTDIR/$GENERATED_DOCKERFILE_NAME $SCRIPTDIR/$A8_SIDECAR_SCRIPT_NAME $SCRIPTDIR/$A8_SIDECAR_TAR_NAME
else
  rm $SCRIPTDIR/$GENERATED_DOCKERFILE_NAME
fi
