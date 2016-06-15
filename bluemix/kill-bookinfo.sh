#!/bin/bash

source .bluemixrc

echo "Listing existing container groups"
EXISTING_GROUPS=$(cf ic group list)

BOOKINFO_GROUPS=(
    bookinfo_productpage
    bookinfo_details
    bookinfo_ratings
    bookinfo_reviews1
    bookinfo_reviews2
    bookinfo_reviews3
    bookinfo_gateway
)

for group in ${BOOKINFO_GROUPS[@]}; do
    echo $EXISTING_GROUPS | grep $group > /dev/null
    if [ $? -eq 0 ]; then
        echo "Removing $group container group"
        cf ic group rm -f $group
    fi
done

echo "Waiting for groups to be removed"
sleep 15
