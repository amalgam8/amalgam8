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
curl -O https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.5.3.linux-amd64.tar.gz
if ! grep -Fq "/home/vagrant/sandbox" /home/vagrant/.profile; then
	echo 'export GOPATH=/home/vagrant/sandbox' >> /home/vagrant/.profile
fi
if ! grep -Fq "/usr/local/go/bin" /home/vagrant/.profile; then
	echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> /home/vagrant/.profile
fi
chown vagrant:vagrant /home/vagrant/sandbox /home/vagrant/sandbox/src /home/vagrant/sandbox/src/github.com

# Install grapnel
cd /home/vagrant
export GOPATH=/home/vagrant/sandbox
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
git clone https://github.com/eanderton/grapnel.git
cd grapnel
make all
ln -sf /home/vagrant/grapnel/bin/grapnel /usr/local/bin/grapnel

# Install docker-compose
sudo curl -L https://github.com/docker/compose/releases/download/1.5.1/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

cat >/tmp/k8s.yml <<EOF
etcd:
  image: gcr.io/google_containers/etcd:2.0.12
  net: "host"
  command: /usr/local/bin/etcd --addr=127.0.0.1:4001 --bind-addr=0.0.0.0:4001 --data-dir=/var/etcd/data
master:
  image: gcr.io/google_containers/hyperkube:v1.1.1
  net: "host"
  pid: "host"
  privileged: true
  ports:
   - "8080:8080"
  volumes:
  - /:/rootfs:ro
  - /sys:/sys:ro
  - /dev:/dev
  - /var/lib/docker/:/var/lib/docker:ro
  - /var/lib/kubelet/:/var/lib/kubelet:rw
  - /var/run:/var/run:rw
  command: /hyperkube kubelet --containerized --hostname-override="127.0.0.1" --address="0.0.0.0" --api-servers=http://0.0.0.0:8080 --enable_server --config=/etc/kubernetes/manifests
proxy:
  image: gcr.io/google_containers/hyperkube:v1.1.1
  net: "host"
  privileged: true
  command: /hyperkube proxy --master=http://0.0.0.0:8080 --v=2
EOF

# Install/Run kubernetes
docker-compose -f /tmp/k8s.yml up -d

# Install kubernetes CLI
sudo curl -L http://storage.googleapis.com/kubernetes-release/release/v1.1.1/bin/linux/amd64/kubectl > /usr/local/bin/kubectl
sudo chmod +x /usr/local/bin/kubectl

##Make API server accessible outside vagrant. Enable if you wish to 
#sleep 5
#docker exec tmp_master_1 perl -pi -e 's/address=127.0.0.1/address=0.0.0.0/' /etc/kubernetes/manifests/master.json
#docker restart tmp_master_1

# Install Amalgam8 CLI

# >>>>>>>>> TODO: when available change the following to: pip install a8ctl
pip install parse
pip install pygremlin
sudo cat >/usr/local/bin/a8ctl <<EOF
#!/bin/sh
python /home/vagrant/sandbox/src/github.com/amalgam8/a8ctl/a8ctl/v1/a8ctl.py \\\$*
EOF
sudo chmod +x /usr/local/bin/a8ctl
# >>>>>>>>>

echo 'export A8_CONTROLLER_URL=http://192.168.33.33:31200' >> /home/vagrant/.profile

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

  # Create a private network, which allows host-only access to the machine
  # using a specific IP.
  config.vm.network "private_network", ip: "192.168.33.33/24"

  config.vm.provision :shell, inline: $script
end
