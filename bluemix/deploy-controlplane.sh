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
  --publish 6379 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --hostname $CONTROLLER_HOSTNAME \
  --domain $ROUTES_DOMAIN \
  --env A8_POLL_INTERVAL=5s \
  --env A8_LOG_LEVEL=debug \
  --env A8_AUTH_MODE=trusted \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${CONTROLLER_IMAGE}

#################################################################################
# Provision Service Discovery, or start a local registry
#################################################################################

if [ "$ENABLE_SERVICEDISCOVERY" = true ]; then
    cf service sd &> /dev/null
    if [ $? -ne 0 ]; then
        echo "Creating a Service Discovery instance..."
        cf create-service service_discovery free sd
    else
        echo "Found an existing Service Discovery instance"
    fi

    if [ $(cf service-key sd sdkey | grep -ic "No service key") -gt 0 ]; then
        echo "Creating Service Discovery credentials"
        cf create-service-key sd sdkey
    else
        echo "Found existing Service Discovery credentials"
    fi

    SDKEY=$(cf service-key sd sdkey | tail -n +3)
    REGISTRY_URL=$(echo "$SDKEY" | jq -r '.url')
    REGISTRY_TOKEN=$(echo "$SDKEY" | jq -r '.auth_token')
else
    echo "Starting registry"
    bluemix ic group-create --name amalgam8_registry \
            --publish 8080 --memory 256 --auto \
            --min 1 --max 2 --desired 1 \
            --hostname $REGISTRY_HOSTNAME \
            --domain $ROUTES_DOMAIN \
            --env AUTH_MODE=trusted \
            ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REGISTRY_IMAGE}
fi

#################################################################################
# Post the local tenant to controller
#################################################################################

# Wait for controller route to set up
echo "Waiting for controller route to set up"
attempt=0
while true; do
    code=$(curl -w "%{http_code}" "${CONTROLLER_URL}/health" -o /dev/null)
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
    code=$(curl -w "%{http_code}" "${REGISTRY_URL}/uptime" -o /dev/null)
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

echo "Setting up a new controller tenant named 'local'"
read -d '' tenant << EOF
{
    "load_balance": "round_robin"
}
EOF

attempt=0
while true; do
    code=$(echo $tenant | curl -w "%{http_code}" -H "Authorization: Bearer local" -H "Content-Type: application/json" -d @- "${CONTROLLER_URL}/v1/tenants")
    if [ "$code" = "201" ]; then
        echo "Controller tenant is set up"
        break
    fi

    attempt=$((attempt + 1))
    if [ "$attempt" -gt 10 ]; then
        echo "Failed setting up controller tenant: controller returned HTTP ${code}"
        echo "Deploying the controlplane has failed"
        exit 1
    fi
    sleep 10s
done

echo "Controlplane has been deployed successfully"
