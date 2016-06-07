#!/bin/bash
export K8S_VERSION="v1.1.1"
# #$(curl -sS https://storage.googleapis.com/kubernetes-release/release/stable.txt)
# docker run \
#     --volume=/:/rootfs:ro \
#     --volume=/sys:/sys:ro \
#     --volume=/var/lib/docker/:/var/lib/docker:rw \
#     --volume=/var/lib/kubelet/:/var/lib/kubelet:rw \
#     --volume=/var/run:/var/run:rw \
#     --net=host \
#     --pid=host \
#     --privileged=true \
#     --name=kubelet \
#     -d \
#     gcr.io/google_containers/hyperkube:${K8S_VERSION} \
#     /hyperkube kubelet \
#         --containerized \
#         --hostname-override="127.0.0.1" \
#         --address="0.0.0.0" \
#         --api-servers=http://localhost:8080 \
#         --config=/etc/kubernetes/manifests \
#         --cluster-dns=10.0.0.10 \
#         --cluster-domain=cluster.local \
#         --allow-privileged=true --v=2

# sudo wget http://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl
# sudo chmod 755 /usr/local/bin/kubectl
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

##Make API server accessible on host OS
sleep 10
docker exec tmp_master_1 perl -pi -e 's/address=127.0.0.1/address=0.0.0.0/' /etc/kubernetes/manifests/master.json
docker restart tmp_master_1

# Install/Run kubernetes
docker-compose -f /tmp/k8s.yml up -d

# Install kubernetes CLI
sudo curl -L http://storage.googleapis.com/kubernetes-release/release/v1.1.1/bin/linux/amd64/kubectl > /usr/local/bin/kubectl
