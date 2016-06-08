#!/bin/bash

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
  command: /hyperkube kubelet --containerized --hostname_override=127.0.0.1 --address=0.0.0.0 --api_servers=http://0.0.0.0:8080 --enable_server --config=/etc/kubernetes/manifests
proxy:
  image: gcr.io/google_containers/hyperkube:v1.1.1
  net: "host"
  privileged: true
  command: /hyperkube proxy --master=http://127.0.0.1:8080 --v=2
EOF

# Install/Run kubernetes
docker-compose -f /tmp/k8s.yml up -d

# Make API server accessible on host OS
sleep 10
docker exec tmp_master_1 sed -i 's/address=127.0.0.1/address=0.0.0.0/' /etc/kubernetes/manifests/master.json
docker restart tmp_master_1

# Install kubernetes CLI
sudo curl -L http://storage.googleapis.com/kubernetes-release/release/v1.1.1/bin/linux/amd64/kubectl > /usr/local/bin/kubectl
