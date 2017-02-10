#!/bin/bash

cleanup_all_rules() {
    if [ "$ENV" == "docker" ]; then
        echo "Cleaning up docker rules"
        $A8CLI rule-delete -a -f
	return $?
    fi

    if [ "$ENV" == "k8s" ]; then
        echo "Cleaning up k8s rules"
        kubectl delete routingrule --all 
	return $?
    fi

    echo "ENV is not set.  Exiting...."
    exit 1
}

create_rule() {
    if [ "$ENV" == "docker" ]; then
	echo $($A8CLI rule-create -f $1)
        return $?
    fi

    if [ "$ENV" == "k8s" ]; then
	echo $(kubectl create -f $1)
        return $?
    fi

    echo "ENV is not set.  Exiting...."
    exit 1
}

list_rules() {
    FIELD=$1
    if [ "$ENV" == "docker" ]; then
        echo $(curl -s -X "GET" $A8_CONTROLLER_URL/v1/rules | jq -r .rules[0].$FIELD)
	return 0
    fi

    if [ "$ENV" == "k8s" ]; then
        echo $(kubectl get routingrule -o json | jq -r .items[0].spec.$FIELD)
	return 0
    fi

    echo "ENV is not set.  Exiting...."
    exit 1
}
