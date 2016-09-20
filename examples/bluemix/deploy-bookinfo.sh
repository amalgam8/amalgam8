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
    echo "$BLUEMIX_IMAGES" | grep "$image" > /dev/null
    if [ $? -ne 0 ]; then
        echo "Pulling ${DOCKERHUB_NAMESPACE}/$image from Dockerhub"
        bluemix ic cpi ${DOCKERHUB_NAMESPACE}/$image ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/$image
    fi
done

#################################################################################
# start the productpage microservice instances
#################################################################################

echo "Starting bookinfo productpage microservice (v1)"

bluemix ic group-create --name bookinfo_productpage \
  --publish 9080 --memory 256 --auto --anti \
  --min 1 --max 2 --desired 1 \
  --env A8_REGISTRY_URL=$REGISTRY_URL \
  --env A8_REGISTRY_POLL=5s \
  --env A8_CONTROLLER_URL=$CONTROLLER_URL \
  --env A8_CONTROLLER_POLL=5s \
  --env A8_SERVICE=productpage:v1 \
  --env A8_ENDPOINT_PORT=9080 \
  --env A8_ENDPOINT_TYPE=http \
  --env A8_REGISTER=true \
  --env A8_PROXY=true \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${PRODUCTPAGE_IMAGE}

#################################################################################
# Start the details microservice instances
#################################################################################

echo "Starting bookinfo details microservice (v1)"

bluemix ic group-create --name bookinfo_details \
  --publish 9080 --memory 256 --auto --anti \
  --min 1 --max 2 --desired 1 \
  --env A8_REGISTRY_URL=$REGISTRY_URL \
  --env A8_SERVICE=details:v1 \
  --env A8_ENDPOINT_PORT=9080 \
  --env A8_ENDPOINT_TYPE=http \
  --env A8_REGISTER=true \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${DETAILS_IMAGE}

#################################################################################
# Start the ratings microservice instances
#################################################################################

echo "Starting bookinfo ratings microservice (v1)"

bluemix ic group-create --name bookinfo_ratings \
  --publish 9080 --memory 256 --auto --anti \
  --min 1 --max 2 --desired 1 \
  --env A8_REGISTRY_URL=$REGISTRY_URL \
  --env A8_SERVICE=ratings:v1 \
  --env A8_ENDPOINT_PORT=9080 \
  --env A8_ENDPOINT_TYPE=http \
  --env A8_REGISTER=true \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${RATINGS_IMAGE}

#################################################################################
# Start the reviews microservice instances
#################################################################################

echo "Starting bookinfo reviews microservice (v1)"

bluemix ic group-create --name bookinfo_reviews1 \
  --publish 9080 --memory 256 --auto --anti \
  --min 1 --max 2 --desired 1 \
  --env A8_REGISTRY_URL=$REGISTRY_URL \
  --env A8_REGISTRY_POLL=5s \
  --env A8_CONTROLLER_URL=$CONTROLLER_URL \
  --env A8_CONTROLLER_POLL=5s \
  --env A8_SERVICE=reviews:v1 \
  --env A8_ENDPOINT_PORT=9080 \
  --env A8_ENDPOINT_TYPE=http \
  --env A8_REGISTER=true \
  --env A8_PROXY=true \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REVIEWS_V1_IMAGE}

echo "Starting bookinfo reviews microservice (v2)"

bluemix ic group-create --name bookinfo_reviews2 \
  --publish 9080 --memory 256 --auto --anti \
  --min 1 --max 2 --desired 1 \
  --env A8_REGISTRY_URL=$REGISTRY_URL \
  --env A8_REGISTRY_POLL=5s \
  --env A8_CONTROLLER_URL=$CONTROLLER_URL \
  --env A8_CONTROLLER_POLL=5s \
  --env A8_SERVICE=reviews:v2 \
  --env A8_ENDPOINT_PORT=9080 \
  --env A8_ENDPOINT_TYPE=http \
  --env A8_REGISTER=true \
  --env A8_PROXY=true \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REVIEWS_V2_IMAGE}

echo "Starting bookinfo reviews microservice (v3)"

bluemix ic group-create --name bookinfo_reviews3 \
  --publish 9080 --memory 256 --auto --anti \
  --min 1 --max 2 --desired 1 \
  --env A8_REGISTRY_URL=$REGISTRY_URL \
  --env A8_REGISTRY_POLL=5s \
  --env A8_CONTROLLER_URL=$CONTROLLER_URL \
  --env A8_CONTROLLER_POLL=5s \
  --env A8_SERVICE=reviews:v3 \
  --env A8_ENDPOINT_PORT=9080 \
  --env A8_ENDPOINT_TYPE=http \
  --env A8_REGISTER=true \
  --env A8_PROXY=true \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/${REVIEWS_V3_IMAGE}

#################################################################################
# Start the gateway
#################################################################################

echo "Starting bookinfo gateway"

bluemix ic group-create --name bookinfo_gateway \
  --publish 6379 --memory 256 --auto --anti \
  --min 1 --max 2 --desired 1 \
  --hostname $BOOKINFO_HOSTNAME \
  --domain $ROUTES_DOMAIN \
  --env A8_REGISTRY_URL=$REGISTRY_URL \
  --env A8_REGISTRY_POLL=5s \
  --env A8_CONTROLLER_URL=$CONTROLLER_URL \
  --env A8_CONTROLLER_POLL=5s \
  --env A8_SERVICE=gateway \
  --env A8_PROXY=true \
  ${BLUEMIX_REGISTRY_HOST}/${BLUEMIX_REGISTRY_NAMESPACE}/$GATEWAY_IMAGE

echo "Bookinfo app has been deployed successfully"
