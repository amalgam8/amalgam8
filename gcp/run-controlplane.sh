#!/bin/bash
set -x

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

rfile="registry.yaml"
cfile="controller.yaml"
mhfile="messagehub.yaml"
lgfile="logserver.yaml"

if [ "$1" == "compile" ]; then
    $SCRIPTDIR/build-sidecar.sh
    if [ $? -ne 0 ]; then
        echo "Sidecar failed to compile"
        exit 1
    fi
    $SCRIPTDIR/build-registry.sh
    if [ $? -ne 0 ]; then
        echo "Registry failed to compile"
        exit 1
    fi
    $SCRIPTDIR/build-controller.sh
    if [ $? -ne 0 ]; then
        echo "Controller failed to compile"
        exit 1
    fi
    exit
elif [ "$1" == "start" ]; then
    echo "Starting Message Hub Local using Kubernertes at kafka:9092"
    kubectl create -f $SCRIPTDIR/$mhfile
    echo "Starting Logmet local using Kubernertes at logserver:8092, elasticsearch at logserver:9200"
    kubectl create -f $SCRIPTDIR/$lgfile
    echo "Starting Registry using Kubernetes..."
    kubectl create -f $SCRIPTDIR/$rfile
    echo "Starting Controller using Kubernetes..."
    kubectl create -f $SCRIPTDIR/$cfile
    echo "Waiting for Controller to initialize..."
    sleep 60
    AR=$(kubectl get svc/registry --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
    AC=$(kubectl get svc/controller --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
    KA=$(kubectl get svc/kafka --template={{.spec.clusterIP}}:{{\("index .spec.ports 0"\).port}})
    echo "Setting up a new tenant named 'local' whose app tracks requests using header 'X-Request-ID'"
    read -d '' tenant << EOF
{
    "id": "local",
    "token": "local",
    "req_tracking_header" : "X-Request-ID",
    "credentials": {
        "kafka": {
            "brokers": ["${KA}"],
            "sasl": false
        },
        "registry": {
            "url": "http://${AR}",
            "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjY3NzU5NjMsIm5hbWVzcGFjZSI6Imdsb2JhbC5nbG9iYWwifQ.Gbz4G_O0OfJZiTuX6Ce4heU83gSWQLr5yyiA7eZNqdY"
        }
    }
}
EOF
    export TENANT_REG=$tenant
    echo "Please assign a public IP to your controller and then issue the following curl command"
    echo 'echo $TENANT_REG|curl -H "Content-Type: application/json" -d @- http://ControllerExternalIP:31200/v1/tenants'
elif [ "$1" == "stop" ]; then
    echo "Stopping Ctrlr, Reg, Lg, and Mh using Kubernetes.."
    kubectl delete -f $SCRIPTDIR/$cfile
    sleep 3
    kubectl delete -f $SCRIPTDIR/$rfile
    sleep 3
    kubectl delete -f $SCRIPTDIR/$lgfile
    sleep 3
    kubectl delete -f $SCRIPTDIR/$mhfile
else
    echo "usage: run-controlplane.sh compile|start|stop"
    exit 1
fi
