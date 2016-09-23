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
# Container group names
#################################################################################

PRODUCT_PAGE_GROUP=bookinfo_productpage
DETAILS_GROUP=bookinfo_details
RATINGS_GROUP=bookinfo_ratings
REVIEWS_V1_GROUP=bookinfo_reviews1
REVIEWS_V2_GROUP=bookinfo_reviews2
REVIEWS_V3_GROUP=bookinfo_reviews3
GATEWAY_GROUP=bookinfo_gateway

BOOKINFO_GROUPS=(
    ${PRODUCT_PAGE_GROUP}
    ${DETAILS_GROUP}
    ${RATINGS_GROUP}
    ${REVIEWS_V1_GROUP}
    ${REVIEWS_V2_GROUP}
    ${REVIEWS_V3_GROUP}
    ${GATEWAY_GROUP}
)

#################################################################################
# start the productpage microservice instances
#################################################################################

echo "Starting bookinfo productpage microservice (v1)"

bluemix ic group-create --name $PRODUCT_PAGE_GROUP \
  --publish 9080 --memory 128 --auto --anti \
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

bluemix ic group-create --name $DETAILS_GROUP \
  --publish 9080 --memory 128 --auto --anti \
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

bluemix ic group-create --name $RATINGS_GROUP \
  --publish 9080 --memory 128 --auto --anti \
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

bluemix ic group-create --name $REVIEWS_V1_GROUP \
  --publish 9080 --memory 128 --auto --anti \
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

bluemix ic group-create --name $REVIEWS_V2_GROUP \
  --publish 9080 --memory 128 --auto --anti \
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

bluemix ic group-create --name $REVIEWS_V3_GROUP \
  --publish 9080 --memory 128 --auto --anti \
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

bluemix ic group-create --name $GATEWAY_GROUP \
  --publish 6379 --memory 128 --auto --anti \
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

#################################################################################
# Check the deployment progress
#################################################################################

echo -e "Waiting for the container groups to be created:"

attempt=0
_wait=true

while $_wait; do
    # Sleep for 15s
    for (( i = 0 ; i < 3 ; i++ )) do
        sleep 5s
        echo -n "."
    done

    EXISTING_GROUPS=$(bluemix ic groups)
    counter=0
    for group in ${BOOKINFO_GROUPS[@]}; do
        status=$(echo "$EXISTING_GROUPS" | awk -v pattern="$group" '$0 ~ pattern { print $3; exit; }')
        case "$status" in
            "CREATE_FAILED")
            _wait=false
        ;;

        "DELETE_FAILED")
            _wait=false
        ;;

        "CREATE_COMPLETE")
            ((counter++))
        ;;
        esac
    done

    if [ "$counter" -eq "${#BOOKINFO_GROUPS[@]}" ]; then
        echo -e "\nBookinfo app has been deployed successfully!"
        break
    fi

    ((attempt++))
    if [ "$attempt" -gt 12 ]; then  # Timeout after 3min
        echo -e "\nTimeout waiting for container groups to be created"
        echo "Deploying bookinfo app has failed"
        exit 1
    fi

    if [[ $_wait = false ]]; then
        echo -e "\nDeploying bookinfo app has failed!\n"
        echo -e "Getting the status of all container groups...\n"
        bluemix ic groups
    fi
done
