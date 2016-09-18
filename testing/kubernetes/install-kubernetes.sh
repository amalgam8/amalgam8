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

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

export K8S_VERSION="v1.2.3"
export ARCH=amd64

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

docker run \
    --volume=/:/rootfs:ro \
    --volume=/sys:/sys:ro \
    --volume=/var/lib/docker/:/var/lib/docker:rw \
    --volume=/var/lib/kubelet/:/var/lib/kubelet:rw \
    --volume=/var/run:/var/run:rw \
    --net=host \
    --pid=host \
    --privileged=true \
    --name=kubelet \
    -d \
    gcr.io/google_containers/hyperkube-${ARCH}:${K8S_VERSION} \
    /hyperkube kubelet \
        --containerized \
        --hostname-override=127.0.0.1 \
        --address=0.0.0.0 \
        --api-servers=http://0.0.0.0:8080 \
        --config=/etc/kubernetes/manifests \
        --allow-privileged=true --v=2 \
  	    --cluster-dns=10.0.0.10 \
        --cluster-domain=cluster.local

# Install kubernetes CLI
curl -L http://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/${ARCH}/kubectl > /tmp/kubectl
sudo mv /tmp/kubectl /usr/local/bin
sudo chmod +x /usr/local/bin/kubectl
echo "Waiting for K8S to initialize.."
sleep 30
