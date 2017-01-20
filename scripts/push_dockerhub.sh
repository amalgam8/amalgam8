#!/bin/bash

###################################################################
### Build, tag and push semantically-versioned images to Docker Hub
###
### This script is intended to be run in a Travis-CI build context,
### where the current commit is tagged with a semantic version tag (e.g. "v3.12.7")
### The following environment variables are assumed to be set:
###
### - DOCKERHUB_EMAIL - Dockerhub login email
### - DOCKERHUB_USERNAME - Docker Hub login username
### - DOCKERHUB_PASSWORD - Docker Hub login password
###############################################################

REGISTRY_IMAGE="amalgam8/a8-registry"
CONTROLLER_IMAGE="amalgam8/a8-controller"
SIDECAR_IMAGE="amalgam8/a8-sidecar"

# Semantic version regular expression (simplified):
SEMVER_REGEX="^v([0-9]+)\.([0-9]+)\.([0-9]+)(-([0-9a-zA-Z.-]+))?$"

echo "Checking for semantic version tag for the current git commit"
tag=$(git describe --exact-match)
if [ $? -ne 0 ]; then
    echo "Skipping deployment to Docker Hub: no git tag found"
    exit 1
fi
echo "Found git tag '$tag'"
if [[ ! $tag =~ $SEMVER_REGEX ]]; then
    echo "Skipping deployment to Docker Hub: git tag is not a semantic version"
    exit 1
fi

patch=$(echo $tag | sed -r "s/$SEMVER_REGEX/\3/")
minor=$(echo $tag | sed -r "s/$SEMVER_REGEX/\2/")
major=$(echo $tag | sed -r "s/$SEMVER_REGEX/\1/")
label=$(echo $tag | sed -r "s/$SEMVER_REGEX/\5/")

if [ ! -z "$label" ]; then
    echo "Skipping deployment to Docker Hub: found a prerelease label ('$label')"
    exit 1
fi
    
patch_tag="$major.$minor.$patch"
minor_tag="$major.$minor"
major_tag="$major"

echo "Tag: $tag"

EXAMPLE_HELLOWORLD_IMAGE_V1="amalgam8/a8-examples-helloworld-v1"
EXAMPLE_HELLOWORLD_IMAGE_V2="amalgam8/a8-examples-helloworld-v2"
EXAMPLE_HELLOWORLD_SIDECAR_V1="amalgam8/a8-examples-helloworld-sidecar-v1"
EXAMPLE_HELLOWORLD_SIDECAR_V2="amalgam8/a8-examples-helloworld-sidecar-v2"
EXAMPLE_BOOKINFO_PRODUCTPAGE_V1="amalgam8/a8-examples-bookinfo-productpage-v1"
EXAMPLE_BOOKINFO_PRODUCTPAGE_SIDECAR_V1="amalgam8/a8-examples-bookinfo-productpage-sidecar-v1"
EXAMPLE_BOOKINFO_DETAILS_V1="amalgam8/a8-examples-bookinfo-details-v1"
EXAMPLE_BOOKINFO_DETAILS_SIDECAR_V1="amalgam8/a8-examples-bookinfo-details-sidecar-v1"
EXAMPLE_BOOKINFO_REVIEWS_V1="amalgam8/a8-examples-bookinfo-reviews-v1"
EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V1="amalgam8/a8-examples-bookinfo-reviews-sidecar-v1"
EXAMPLE_BOOKINFO_REVIEWS_V2="amalgam8/a8-examples-bookinfo-reviews-v2"
EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V2="amalgam8/a8-examples-bookinfo-reviews-sidecar-v2"
EXAMPLE_BOOKINFO_REVIEWS_V3="amalgam8/a8-examples-bookinfo-reviews-v3"
EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V3="amalgam8/a8-examples-bookinfo-reviews-sidecar-v3"
EXAMPLE_BOOKINFO_RATINGS_V1="amalgam8/a8-examples-bookinfo-ratings-v1"
EXAMPLE_BOOKINFO_RATINGS_SIDECAR_V1="amalgam8/a8-examples-bookinfo-ratings-sidecar-v1"

echo "Building docker images..."
docker build -t "$REGISTRY_IMAGE:latest" -f docker/Dockerfile.registry .
docker build -t "$CONTROLLER_IMAGE:latest" -f docker/Dockerfile.controller .
docker build -t "$SIDECAR_IMAGE:latest" -f docker/Dockerfile.sidecar.envoy.ubuntu .

echo "Listing current image tags in Docker Hub..."
dockerhub_tags=$(curl --silent "https://registry.hub.docker.com/v1/repositories/$REGISTRY_IMAGE/tags" | jq -r ".[].name")

# Always push the patch version tag (e.g., '3.12.7')
push_patch=true

# Determine if the minor tag (e.g., '3.12') should be pushed
max_patch=$(echo "$dockerhub_tags" | sed -rn "s/$minor_tag\.([0-9]+)/\1/p" | sort -r | head -n1)
if [[ -z "$max_patch" || $patch -ge $max_patch ]]; then
    push_minor=true
fi

# Determine if the major tag (e.g., '3') should be pushed
max_minor=$(echo "$dockerhub_tags" | sed -rn "s/$major_tag\.([0-9]+)\.[0-9]+/\1/p" | sort -r | head -n1)
if [[ $major -gt 0 && $push_minor = true && ( -z "$max_minor" || $minor -ge $max_minor ) ]]; then
    push_major=true
fi

# Determine if the 'latest' tag should be pushed
max_major=$(echo "$dockerhub_tags" | sed -rn "s/([0-9]+)\.[0-9]+\.[0-9]+/\1/p" | sort -r | head -n1)
if [[ ( $push_major = true || $major -eq 0 ) && $push_minor = true && ( -z "$max_major" || $major -ge $max_major ) ]]; then
    push_latest=true
fi

echo "Signing into Docker Hub..."
docker login --email $DOCKERHUB_EMAIL --username $DOCKERHUB_USERNAME --password $DOCKERHUB_PASSWORD

if [ "$push_patch" = true ]; then
    echo "Pushing '$REGISTRY_IMAGE:$patch_tag' to Docker Hub..."
    docker tag "$REGISTRY_IMAGE:latest" "$REGISTRY_IMAGE:$patch_tag"
    docker push "$REGISTRY_IMAGE:$patch_tag"
    
    echo "Pushing '$CONTROLLER_IMAGE:$patch_tag' to Docker Hub..."
    docker tag "$CONTROLLER_IMAGE:latest" "$CONTROLLER_IMAGE:$patch_tag"
    docker push "$CONTROLLER_IMAGE:$patch_tag"
    
    echo "Pushing '$SIDECAR_IMAGE:$patch_tag' to Docker Hub..."
    docker tag "$SIDECAR_IMAGE:latest" "$SIDECAR_IMAGE:$patch_tag"
    docker push "$SIDECAR_IMAGE:$patch_tag"
    
    echo "Pushing '$SIDECAR_IMAGE:$patch_tag-alpine' to Docker Hub..."
    docker tag "$SIDECAR_IMAGE:alpine" "$SIDECAR_IMAGE:$patch_tag-alpine"
    docker push "$SIDECAR_IMAGE:$patch_tag-alpine"

    echo "Pushing '$EXAMPLE_HELLOWORLD_IMAGE_V1' to Docker Hub..."
    docker push "$EXAMPLE_HELLOWORLD_IMAGE_V1:$patch_tag"

    echo "Pushing '$EXAMPLE_HELLOWORLD_SIDECAR_V1' to Docker Hub..."
    docker push "$EXAMPLE_HELLOWORLD_SIDECAR_V1:$patch_tag"

    echo "Pushing '$EXAMPLE_HELLOWORLD_IMAGE_V2' to Docker Hub..."
    docker push "$EXAMPLE_HELLOWORLD_IMAGE_V2:$patch_tag"

    echo "Pushing '$EXAMPLE_HELLOWORLD_SIDECAR_V2' to Docker Hub..."
    docker push "$EXAMPLE_HELLOWORLD_SIDECAR_V2:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_PRODUCTPAGE_V1' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_PRODUCTPAGE_V1:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_PRODUCTPAGE_SIDECAR_V1' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_PRODUCTPAGE_SIDECAR_V1:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_DETAILS_V1' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_DETAILS_V1:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_DETAILS_SIDECAR_V1' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_DETAILS_SIDECAR_V1:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_V1' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_V1:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V1' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V1:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_V2' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_V2:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V2' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V2:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_V3' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_V3:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V3' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V3:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_RATINGS_V1' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_RATINGS_V1:$patch_tag"

    echo "Pushing '$EXAMPLE_BOOKINFO_RATINGS_SIDECAR_V1' to Docker Hub..."
    docker push "$EXAMPLE_BOOKINFO_RATINGS_SIDECAR_V1:$patch_tag"
>>>>>>> Push example images to dockerhub when creating a release.
fi
if [ "$push_minor" = true ]; then
    echo "Pushing '$REGISTRY_IMAGE:$minor_tag' to Docker Hub..."
    docker tag "$REGISTRY_IMAGE:latest" "$REGISTRY_IMAGE:$minor_tag"
    docker push "$REGISTRY_IMAGE:$minor_tag"
    
    echo "Pushing '$CONTROLLER_IMAGE:$minor_tag' to Docker Hub..."
    docker tag "$CONTROLLER_IMAGE:latest" "$CONTROLLER_IMAGE:$minor_tag"
    docker push "$CONTROLLER_IMAGE:$minor_tag"
    
    echo "Pushing '$SIDECAR_IMAGE:$minor_tag' to Docker Hub..."
    docker tag "$SIDECAR_IMAGE:latest" "$SIDECAR_IMAGE:$minor_tag"
    docker push "$SIDECAR_IMAGE:$minor_tag"
    
    # echo "Pushing '$SIDECAR_IMAGE:$minor_tag-alpine' to Docker Hub..."
    # docker tag "$SIDECAR_IMAGE:alpine" "$SIDECAR_IMAGE:$minor_tag-alpine"
    # docker push "$SIDECAR_IMAGE:$minor_tag-alpine"
fi
if [ "$push_major" = true ]; then
    echo "Pushing '$REGISTRY_IMAGE:$major_tag' to Docker Hub..."
    docker tag "$REGISTRY_IMAGE:latest" "$REGISTRY_IMAGE:$major_tag"
    docker push "$REGISTRY_IMAGE:$major_tag"
    
    echo "Pushing '$CONTROLLER_IMAGE:$major_tag' to Docker Hub..."
    docker tag "$CONTROLLER_IMAGE:latest" "$CONTROLLER_IMAGE:$major_tag"
    docker push "$CONTROLLER_IMAGE:$major_tag"
    
    echo "Pushing '$SIDECAR_IMAGE:$major_tag' to Docker Hub..."
    docker tag "$SIDECAR_IMAGE:latest" "$SIDECAR_IMAGE:$major_tag"
    docker push "$SIDECAR_IMAGE:$major_tag"
    
    # echo "Pushing '$SIDECAR_IMAGE:$major_tag-alpine' to Docker Hub..."
    # docker tag "$SIDECAR_IMAGE:alpine" "$SIDECAR_IMAGE:$major_tag-alpine"
    # docker push "$SIDECAR_IMAGE:$major_tag-alpine"
fi
if [ "$push_latest" = true ]; then
    echo "Pushing '$REGISTRY_IMAGE:latest' to Docker Hub..."
    docker push "$REGISTRY_IMAGE:latest"
    
    echo "Pushing '$CONTROLLER_IMAGE:latest' to Docker Hub..."
    docker push "$CONTROLLER_IMAGE:latest"
    
    echo "Pushing '$SIDECAR_IMAGE:latest' to Docker Hub..."
    docker push "$SIDECAR_IMAGE:latest"
    
    echo "Pushing '$SIDECAR_IMAGE:alpine' to Docker Hub..."
    docker push "$SIDECAR_IMAGE:alpine"

    echo "Pushing '$EXAMPLE_HELLOWORLD_IMAGE_V1' to Docker Hub..."
    docker tag "$EXAMPLE_HELLOWORLD_IMAGE_V1:$patch_tag" "$EXAMPLE_HELLOWORLD_IMAGE_V1:latest"
    docker push "$EXAMPLE_HELLOWORLD_IMAGE_V1"

    echo "Pushing '$EXAMPLE_HELLOWORLD_SIDECAR_V1' to Docker Hub..."
    docker tag "$EXAMPLE_HELLOWORLD_SIDECAR_V1:$patch_tag" "$EXAMPLE_HELLOWORLD_SIDECAR_V1:latest"
    docker push "$EXAMPLE_HELLOWORLD_SIDECAR_V1"

    echo "Pushing '$EXAMPLE_HELLOWORLD_IMAGE_V2' to Docker Hub..."
    docker tag "$EXAMPLE_HELLOWORLD_IMAGE_V2:$patch_tag" "$EXAMPLE_HELLOWORLD_IMAGE_V2:latest"
    docker push "$EXAMPLE_HELLOWORLD_IMAGE_V2"

    echo "Pushing '$EXAMPLE_HELLOWORLD_SIDECAR_V2' to Docker Hub..."
    docker tag "$EXAMPLE_HELLOWORLD_SIDECAR_V2:$patch_tag" "$EXAMPLE_HELLOWORLD_SIDECAR_V2:latest"
    docker push "$EXAMPLE_HELLOWORLD_SIDECAR_V2"

    echo "Pushing '$EXAMPLE_BOOKINFO_PRODUCTPAGE_V1' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_PRODUCTPAGE_V1:$patch_tag" "$EXAMPLE_BOOKINFO_PRODUCTPAGE_V1:latest"
    docker push "$EXAMPLE_BOOKINFO_PRODUCTPAGE_V1"

    echo "Pushing '$EXAMPLE_BOOKINFO_PRODUCTPAGE_SIDECAR_V1' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_PRODUCTPAGE_SIDECAR_V1:$patch_tag" "$EXAMPLE_BOOKINFO_PRODUCTPAGE_SIDECAR_V1:latest"
    docker push "$EXAMPLE_BOOKINFO_PRODUCTPAGE_SIDECAR_V1"

    echo "Pushing '$EXAMPLE_BOOKINFO_DETAILS_V1' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_DETAILS_V1:$patch_tag" "$EXAMPLE_BOOKINFO_DETAILS_V1:latest"
    docker push "$EXAMPLE_BOOKINFO_DETAILS_V1"

    echo "Pushing '$EXAMPLE_BOOKINFO_DETAILS_SIDECAR_V1' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_DETAILS_SIDECAR_V1:$patch_tag" "$EXAMPLE_BOOKINFO_DETAILS_SIDECAR_V1:latest"
    docker push "$EXAMPLE_BOOKINFO_DETAILS_SIDECAR_V1"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_V1' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_REVIEWS_V1:$patch_tag" "$EXAMPLE_BOOKINFO_REVIEWS_V1:latest"
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_V1"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V1' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V1:$patch_tag" "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V1:latest"
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V1"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_V2' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_REVIEWS_V2:$patch_tag" "$EXAMPLE_BOOKINFO_REVIEWS_V2:latest"
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_V2"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V2' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V2:$patch_tag" "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V2:latest"
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V2"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_V3' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_REVIEWS_V3:$patch_tag" "$EXAMPLE_BOOKINFO_REVIEWS_V3:latest"
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_V3"

    echo "Pushing '$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V3' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V3:$patch_tag" "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V3:latest"
    docker push "$EXAMPLE_BOOKINFO_REVIEWS_SIDECAR_V3"

    echo "Pushing '$EXAMPLE_BOOKINFO_RATINGS_V1' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_RATINGS_V1:$patch_tag" "$EXAMPLE_BOOKINFO_RATINGS_V1:latest"
    docker push "$EXAMPLE_BOOKINFO_RATINGS_V1"

    echo "Pushing '$EXAMPLE_BOOKINFO_RATINGS_SIDECAR_V1' to Docker Hub..."
    docker tag "$EXAMPLE_BOOKINFO_RATINGS_SIDECAR_V1:$patch_tag" "$EXAMPLE_BOOKINFO_RATINGS_SIDECAR_V1:latest"
    docker push "$EXAMPLE_BOOKINFO_RATINGS_SIDECAR_V1"
>>>>>>> Push example images to dockerhub when creating a release.
fi

echo "Signing out of Docker Hub"
docker logout
