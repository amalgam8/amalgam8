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

# This Vagrantfile starts a Ubuntu machine sandbox environment with the following installed and running:
#   1. Docker
#   2. docker-compose
#   3. kubernetes

# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<SCRIPT
set -x

apt-get update -qq
apt-get install -q -y curl python-pip jq

# Install and run Docker
echo deb http://get.docker.com/ubuntu docker main > /etc/apt/sources.list.d/docker.list
apt-key adv --keyserver pgp.mit.edu --recv-keys 36A1D7869245C8950F966E92D8576A8BA88D21E9
apt-get update
sudo wget -qO- https://get.docker.com/ | sh

sudo usermod -a -G docker vagrant # Add vagrant user to the docker group

sudo cat >/usr/local/bin/denter <<EOF
#!/bin/sh
docker exec -it \\\$1 bash
EOF
sudo chmod +x /usr/local/bin/denter

# Install golang
cd /tmp
curl -O https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz
if ! grep -Fq "/home/vagrant/sandbox" /home/vagrant/.profile; then
	echo 'export GOPATH=/home/vagrant/sandbox' >> /home/vagrant/.profile
fi
if ! grep -Fq "/usr/local/go/bin" /home/vagrant/.profile; then
	echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> /home/vagrant/.profile
fi
chown vagrant:vagrant /home/vagrant/sandbox /home/vagrant/sandbox/src /home/vagrant/sandbox/src/github.com

# Install docker-compose
sudo curl -L https://github.com/docker/compose/releases/download/1.5.1/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

## Install kubernetes
#export K8S_VERSION="v1.2.3"
#export ARCH=amd64
#docker run \
#    --volume=/:/rootfs:ro \
#    --volume=/sys:/sys:ro \
#    --volume=/var/lib/docker/:/var/lib/docker:rw \
#    --volume=/var/lib/kubelet/:/var/lib/kubelet:rw \
#    --volume=/var/run:/var/run:rw \
#    --net=host \
#    --pid=host \
#    --privileged=true \
#    --name=kubelet \
#    -d \
#    gcr.io/google_containers/hyperkube-${ARCH}:${K8S_VERSION} \
#    /hyperkube kubelet \
#        --containerized \
#        --hostname-override="127.0.0.1" \
#        --address="0.0.0.0" \
#        --api-servers=http://0.0.0.0:8080 \
#        --config=/etc/kubernetes/manifests \
#        --allow-privileged=true --v=2 \
#  	    --cluster-dns=10.0.0.10 \
#        --cluster-domain=cluster.local
## ##Make API server accessible on host OS
#sleep 10
#docker exec kubelet perl -pi -e 's/address=127.0.0.1/address=0.0.0.0/' /etc/kubernetes/manifests/master.json
#docker restart kubelet

## Install kubernetes CLI
#sudo curl -L http://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/${ARCH}/kubectl > /usr/local/bin/kubectl
#sudo chmod +x /usr/local/bin/kubectl

##Make API server accessible outside vagrant. Enable if you wish to 
#sleep 5
#docker exec tmp_master_1 perl -pi -e 's/address=127.0.0.1/address=0.0.0.0/' /etc/kubernetes/manifests/master.json
#docker restart tmp_master_1

# Install Amalgam8 CLI
pip install a8ctl==0.1.8

## Installing Weave Scope
#sudo wget -O /usr/local/bin/scope https://git.io/scope
#sudo chmod a+x /usr/local/bin/scope
###Don't launch it yet.
##sudo scope launch

#echo 'export A8_CONTROLLER_URL=http://192.168.33.33:31200' >> /home/vagrant/.profile

SCRIPT

Vagrant.configure('2') do |config|
  config.vm.box = "ubuntu/trusty64"

  config.vm.synced_folder ".", "/vagrant", disabled: true

  config.vm.synced_folder ".", "/home/vagrant/sandbox/src/github.com/amalgam8/examples"

  if FileTest::directory?("../sidecar")
    config.vm.synced_folder "../sidecar", "/home/vagrant/sandbox/src/github.com/amalgam8/sidecar"
  end

  if FileTest::directory?("../controller")
    config.vm.synced_folder "../controller", "/home/vagrant/sandbox/src/github.com/amalgam8/controller"
  end

  if FileTest::directory?("../registry")
    config.vm.synced_folder "../registry", "/home/vagrant/sandbox/src/github.com/amalgam8/registry"
  end

  if FileTest::directory?("../a8ctl")
    config.vm.synced_folder "../a8ctl", "/home/vagrant/sandbox/src/github.com/amalgam8/a8ctl"
  end

  config.vm.provider :virtualbox do |vb|
    #vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
    vb.customize ['modifyvm', :id, '--memory', '4096']
    vb.cpus = 2
  end

  # Port mappings for various services inside the VM
  ####Controller
  config.vm.network "forwarded_port", guest: 31200, host: 31200
  ####Registry
  config.vm.network "forwarded_port", guest: 31300, host: 31300
  ####Gateway
  config.vm.network "forwarded_port", guest: 32000, host: 32000
  ####Elasticsearch
  config.vm.network "forwarded_port", guest: 30200, host: 30200
  ####Kibana
  config.vm.network "forwarded_port", guest: 30500, host: 30500
  ####Weave Scope
  config.vm.network "forwarded_port", guest: 30040, host: 30040
  ####Marathon Dashboard/Kubernetes dashboard
  config.vm.network "forwarded_port", guest: 8080, host: 38080
  ####Mesos dashboard
  config.vm.network "forwarded_port", guest: 35050, host: 35050

  # Create a private network, which allows host-only access to the machine
  # using a specific IP. For Marathon only.
  #config.vm.network "private_network", ip: "192.168.33.33/24"

  config.vm.provision :shell, inline: $script
end
