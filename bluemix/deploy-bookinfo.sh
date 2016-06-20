#!/bin/bash

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source $SCRIPTDIR/.bluemixrc

#################################################################################
# Pull Dockerhub images
#################################################################################

echo "Looking up Bluemix registry images"
BLUEMIX_IMAGES=$(bluemix ic images --format "{{.Repository}}:{{.Tag}}")

REQUIRED_IMAGES=(
    ${PRODUCTPAGE_IMAGE}
    ${DETAILS_IMAGE}
    ${RATINGS_IMAGE}
    ${REVIEWS_V1_IMAGE}
    ${REVIEWS_V2_IMAGE}
    ${REVIEWS_V3_IMAGE}
    ${GATEWAY_IMAGE}
)

for image in ${REQUIRED_IMAGES[@]}; do
    echo $BLUEMIX_IMAGES | grep $image > /dev/null
    if [ $? -ne 0 ]; then
        echo "Pulling ${DOCKERHUB_NAMESPACE}/$image from Dockerhub"
        bluemix ic cpi ${DOCKERHUB_NAMESPACE}/$image ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/$image
    fi
done

#################################################################################
# Fetch registry credentials
#################################################################################

if [ "$ENABLE_SERVICEDISCOVERY" = true ]; then
    SDKEY=$(cf service-key sd sdkey | tail -n +3)
    REGISTRY_URL=$(echo "$SDKEY" | jq -r '.url')
    REGISTRY_TOKEN=$(echo "$SDKEY" | jq -r '.auth_token')
else
    # TODO: Use local registry credentials
    echo "Not not implemented"
    exit 1
fi

#################################################################################
# Start the productpage microservice instances
#################################################################################

echo "Starting bookinfo productpage microservice (v1)"

bluemix ic group-create --name bookinfo_productpage \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_ID=$CONTROLLER_TENANT_ID \
  --env TENANT_TOKEN=$CONTROLLER_TENANT_TOKEN \
  --env SERVICE=productpage \
  --env SERVICE_VERSION=v1 \
  --env ENDPOINT_PORT=9080 \
  --env LOG=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${PRODUCTPAGE_IMAGE}:v1 

#################################################################################
# Start the details microservice instances
#################################################################################

echo "Starting bookinfo details microservice (v1)"

bluemix ic group-create --name bookinfo_details \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env REGISTRY_URL=$REGISTRY_URL \
  --env REGISTRY_TOKEN=$REGISTRY_TOKEN \
  --env SERVICE=details \
  --env SERVICE_VERSION=v1 \
  --env ENDPOINT_PORT=9080 \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${DETAILS_IMAGE}

#################################################################################
# Start the ratings microservice instances
#################################################################################

echo "Starting bookinfo ratings microservice (v1)"

bluemix ic group-create --name bookinfo_ratings \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env REGISTRY_URL=$REGISTRY_URL \
  --env REGISTRY_TOKEN=$REGISTRY_TOKEN \
  --env SERVICE=ratings \
  --env SERVICE_VERSION=v1 \
  --env ENDPOINT_PORT=9080 \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${RATINGS_IMAGE}
    
#################################################################################
# Start the reviews microservice instances
#################################################################################

echo "Starting bookinfo reviews microservice (v1)"

bluemix ic group-create --name bookinfo_reviews1 \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_ID=$CONTROLLER_TENANT_ID \
  --env TENANT_TOKEN=$CONTROLLER_TENANT_TOKEN \
  --env SERVICE=reviews \
  --env SERVICE_VERSION=v1 \
  --env ENDPOINT_PORT=9080 \
  --env LOG=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REVIEWS_V1_IMAGE}

echo "Starting bookinfo reviews microservice (v2)"

bluemix ic group-create --name bookinfo_reviews2 \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_ID=$CONTROLLER_TENANT_ID \
  --env TENANT_TOKEN=$CONTROLLER_TENANT_TOKEN \
  --env SERVICE=reviews \
  --env SERVICE_VERSION=v2 \
  --env ENDPOINT_PORT=9080 \
  --env LOG=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REVIEWS_V2_IMAGE}

echo "Starting bookinfo reviews microservice (v3)"

bluemix ic group-create --name bookinfo_reviews3 \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_ID=$CONTROLLER_TENANT_ID \
  --env TENANT_TOKEN=$CONTROLLER_TENANT_TOKEN \
  --env SERVICE=reviews \
  --env SERVICE_VERSION=v3 \
  --env ENDPOINT_PORT=9080 \
  --env LOG=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REVIEWS_V3_IMAGE}
    
#################################################################################
# Start the gateway
#################################################################################

echo "Starting bookinfo gateway"

bluemix ic group-create --name bookinfo_gateway \
  --publish 6379 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_ID=$CONTROLLER_TENANT_ID \
  --env TENANT_TOKEN=$CONTROLLER_TENANT_TOKEN \
  --env SERVICE=gateway \
  --env SERVICE_VERSION=v1 \
  --env LOG=false \
  --env REGISTER=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/$GATEWAY_IMAGE

echo "Waiting for gateway to start..."​
sleep 15s

echo "Mapping route to gateway: $BOOKINFO_URL"
bluemix ic route-map --hostname $BOOKINFO_HOSTNAME --domain $ROUTES_DOMAIN bookinfo_gateway

