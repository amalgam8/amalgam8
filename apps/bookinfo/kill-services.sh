#!/bin/bash
#
# Copyright 2016 IBM Corporation
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

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

#gunicorn -D -w 1 -b 0.0.0.0:10081 --reload details:app
#gunicorn -D -w 1 -b 0.0.0.0:10082 --reload reviews:app
#gunicorn -w 1 -b 0.0.0.0:19080 --reload --access-logfile prod.log --error-logfile prod.log productpage:app >>prod.log 2>&1 &
kubectl delete -f $SCRIPTDIR/bookinfo.yaml
