#!/bin/bash

source .bluemixrc

#################################################################################
# Pull Dockerhub images
#################################################################################

echo "Looking up Bluemix registry images"
BLUEMIX_IMAGES=$(cf ic images | tail -n +2 | sed -r 's/([^ ]+) +([^ ]+).*/\1:\2/')

REQUIRED_IMAGES=(
    ${PRODUCTPAGE_IMAGE}:v1
    ${DETAILS_IMAGE}:v1
    ${RATINGS_IMAGE}:v1
    ${REVIEWS_IMAGE}:v1
    ${REVIEWS_IMAGE}:v2
    ${REVIEWS_IMAGE}:v3
    ${GATEWAY_IMAGE}
)

for image in ${REQUIRED_IMAGES[@]}; do
    echo $BLUEMIX_IMAGES | grep $image > /dev/null
    if [ $? -ne 0 ]; then
        echo "Pulling ${DOCKERHUB_NAMESPACE}/$image from Dockerhub"
        cf ic cpi ${DOCKERHUB_NAMESPACE}/$image ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/$image
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

cf ic group create --name bookinfo_productpage \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_ID=$CONTROLLER_TENANT_ID \
  --env TENANT_TOKEN=$CONTROLLER_TENANT_TOKEN \
  --env SERVICE=productpage:v1 \
  --env ENDPOINT_PORT=9080 \
  --env LOG=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${PRODUCTPAGE_IMAGE}:v1 

#################################################################################
# Start the details microservice instances
#################################################################################

echo "Starting bookinfo details microservice (v1)"

cf ic group create --name bookinfo_details \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env REGISTRY_URL=$REGISTRY_URL \
  --env REGISTRY_TOKEN=$REGISTRY_TOKEN \
  --env SERVICE=details:v1 \
  --env ENDPOINT_PORT=9080 \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${DETAILS_IMAGE}:v1

#################################################################################
# Start the ratings microservice instances
#################################################################################

echo "Starting bookinfo ratings microservice (v1)"

cf ic group create --name bookinfo_ratings \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env REGISTRY_URL=$REGISTRY_URL \
  --env REGISTRY_TOKEN=$REGISTRY_TOKEN \
  --env SERVICE=ratings:v1 \
  --env ENDPOINT_PORT=9080 \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${RATINGS_IMAGE}:v1
    
#################################################################################
# Start the reviews microservice instances
#################################################################################

echo "Starting bookinfo reviews microservice (v1)"

cf ic group create --name bookinfo_reviews1 \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_ID=$CONTROLLER_TENANT_ID \
  --env TENANT_TOKEN=$CONTROLLER_TENANT_TOKEN \
  --env REGISTRY_URL=$REGISTRY_URL \
  --env REGISTRY_TOKEN=$REGISTRY_TOKEN \
  --env SERVICE=reviews:v1 \
  --env ENDPOINT_PORT=9080 \
  --env LOG=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REVIEWS_IMAGE}:v1

echo "Starting bookinfo reviews microservice (v2)"

cf ic group create --name bookinfo_reviews2 \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_ID=$CONTROLLER_TENANT_ID \
  --env TENANT_TOKEN=$CONTROLLER_TENANT_TOKEN \
  --env REGISTRY_URL=$REGISTRY_URL \
  --env REGISTRY_TOKEN=$REGISTRY_TOKEN \
  --env SERVICE=reviews:v2 \
  --env ENDPOINT_PORT=9080 \
  --env LOG=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REVIEWS_IMAGE}:v2

echo "Starting bookinfo reviews microservice (v3)"

cf ic group create --name bookinfo_reviews3 \
  --publish 9080 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_ID=$CONTROLLER_TENANT_ID \
  --env TENANT_TOKEN=$CONTROLLER_TENANT_TOKEN \
  --env REGISTRY_URL=$REGISTRY_URL \
  --env REGISTRY_TOKEN=$REGISTRY_TOKEN \
  --env SERVICE=reviews:v3 \
  --env ENDPOINT_PORT=9080 \
  --env LOG=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REVIEWS_IMAGE}:v3
    
#################################################################################
# Start the gateway
#################################################################################

echo "Starting bookinfo gateway"

cf ic group create --name bookinfo_gateway \
  --publish 80 --memory 256 --auto \
  --min 1 --max 2 --desired 1 \
  --env CONTROLLER_URL=$CONTROLLER_URL \
  --env TENANT_TOKEN=12345 \
  --env TENANT_ID=local \
  --env LOG=false \
  --env REGISTER=false \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/$GATEWAY_IMAGE

echo "Waiting for gateway to start..."​
sleep 15s

echo "Mapping route to gateway: $BOOKINFO_URL"
cf ic route map --hostname $BOOKINFO_HOSTNAME --domain $ROUTES_DOMAIN bookinfo_gateway

