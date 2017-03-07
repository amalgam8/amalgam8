#!/bin/bash
#
# Copyright 2017 IBM Corporation
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

cleanup_all_rules() {
    echo "Cleaning up rules"
    $A8CLI rule-delete -a -f
	return $?
}

create_rule() {
	echo $($A8CLI rule-create -f $1)
    return $?
}

list_rules() {
    FIELD=$1
    RULECOUNT=$2
    echo $(curl -s -X "GET" $A8_CONTROLLER_URL/v1/rules | jq -r .rules[$RULECOUNT].$FIELD)
	return 0
}

delete_rule() {
    $A8CLI rule-delete -i $1
    return $?
}
