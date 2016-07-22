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
    echo $BLUEMIX_IMAGES | grep $image > /dev/null
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
            ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REGISTRY_IMAGE} -- /opt/a8registry/a8registry -auth_mode=trusted
fi

#################################################################################
# Provision Message Hub, or fallback to polling
#################################################################################

if [ "$ENABLE_MESSAGEHUB" = true ]; then
    cf service mh &> /dev/null
    if [ $? -ne 0 ]; then
        echo "Creating a Message Hub instance..."
        cf create-service messagehub standard mh
    else
        echo "Found an existing Message Hub instance"
    fi

    if [ $(cf service-key mh mhkey | grep -ic "No service key") -gt 0 ]; then
        echo "Creating Message Hub credentials"
        cf create-service-key mh mhkey
    else
        echo "Found existing Message Hub credentials"
    fi

    MHKEY=$(cf service-key mh mhkey | tail -n +3)
    KAFKA_API_KEY=$(echo "$MHKEY" | jq -r '.api_key')
    KAFKA_ADMIN_URL=$(echo "$MHKEY" | jq -r '.kafka_admin_url')
    KAFKA_REST_URL=$(echo "$MHKEY" | jq -r '.kafka_rest_url')
    KAFKA_BROKERS=$(echo "$MHKEY" | jq -r '.kafka_brokers_sasl')
    KAFKA_USER=$(echo "$MHKEY" | jq -r '.user')
    KAFKA_PASSWORD=$(echo "$MHKEY" | jq -r '.password')
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
        echo "Timeout waiting for controller route..."
        echo "Deploying the controlplane has failed"
        exit 1
    fi
    sleep 10s
done

# Wait for registry route to set up
echo "Waiting for registry route to set up"
attempt=0
while true; do
    code=$(curl -w "%{http_code}" "${REGISTRY_URL}/health" -o /dev/null)
    if [ "$code" = "200" ]; then
        echo "Registry route is set to '$REGISTRY_URL'"
        break
    fi

    attempt=$((attempt + 1))
    if [ "$attempt" -gt 10 ]; then
        echo "Timeout waiting for registry route..."
        echo "Deploying the controlplane has failed"
        exit 1
    fi
    sleep 10s
done

echo "Setting up a new controller tenant named 'local'"
read -d '' tenant << EOF
{
    "credentials": {
        "registry": {
            "url": "${REGISTRY_URL}",
            "token": "${REGISTRY_TOKEN}"
        }
    }
}
EOF

if [ "$ENABLE_MESSAGEHUB" = true ]; then
    read -d '' kafka << EOF
{
    "credentials": {
        "kafka": {
            "api_key": "${KAFKA_API_KEY}",
            "admin_url": "${KAFKA_ADMIN_URL}",
            "rest_url": "${KAFKA_REST_URL}",
            "brokers": ${KAFKA_BROKERS},
            "user": "${KAFKA_USER}",
            "password": "${KAFKA_PASSWORD}",
            "sasl": true
        }
    }
}
EOF
    tenant=$(jq -s '.[0] * .[1]' <(echo $tenant) <(echo $kafka))
fi

echo $tenant | curl -H "Content-Type: application/json" -d @- "${CONTROLLER_URL}/v1/tenants"
