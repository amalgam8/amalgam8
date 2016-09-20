#!/bin/bash

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source $SCRIPTDIR/.bluemixrc

#################################################################################
# Pull Dockerhub images
#################################################################################

echo "Looking up Bluemix registry images"
BLUEMIX_IMAGES=$(bluemix ic images --format "{{.Repository}}:{{.Tag}}")

REQUIRED_IMAGES=(
    ${CONTROLLER_IMAGE}
    ${REGISTRY_IMAGE}
)

for image in ${REQUIRED_IMAGES[@]}; do
    echo "$BLUEMIX_IMAGES" | grep "$image" > /dev/null
    if [ $? -ne 0 ]; then
        echo "Pulling ${DOCKERHUB_NAMESPACE}/$image from Dockerhub"
        bluemix ic cpi ${DOCKERHUB_NAMESPACE}/$image ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/$image
    fi
done

#################################################################################
# Start a controller and registry
#################################################################################

echo "Starting controller"
bluemix ic group-create --name amalgam8_controller \
  --publish 8080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --hostname $CONTROLLER_HOSTNAME \
  --domain $ROUTES_DOMAIN \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${CONTROLLER_IMAGE}

echo "Starting registry"
bluemix ic group-create --name amalgam8_registry \
  --publish 8080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --hostname $REGISTRY_HOSTNAME \
  --domain $ROUTES_DOMAIN \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REGISTRY_IMAGE}

# Wait for controller route to set up
echo "Waiting for controller route to set up"
attempt=0
while true; do
    code=$(curl -w "%{http_code}" --max-time 10 "${CONTROLLER_URL}/health" -o /dev/null)
    if [ "$code" = "200" ]; then
        echo "Controller route is set to '$CONTROLLER_URL'"
        break
    fi

    attempt=$((attempt + 1))
    if [ "$attempt" -gt 10 ]; then
        echo "Timeout waiting for controller route: /health returned HTTP ${code}"
        echo "Deploying the controlplane has failed"
        exit 1
    fi
    sleep 10s
done

# Wait for registry route to set up
echo "Waiting for registry route to set up"
attempt=0
while true; do
    code=$(curl -w "%{http_code}" --max-time 10 "${REGISTRY_URL}/uptime" -o /dev/null)
    if [ "$code" = "200" ]; then
        echo "Registry route is set to '$REGISTRY_URL'"
        break
    fi

    attempt=$((attempt + 1))
    if [ "$attempt" -gt 10 ]; then
        echo "Timeout waiting for registry route: /uptime returned HTTP ${code}"
        echo "Deploying the controlplane has failed"
        exit 1
    fi
    sleep 10s
done

echo "Controlplane has been deployed successfully"
