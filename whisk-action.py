#!/usr/bin/python
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

import requests
import json
import sys
def main(dict):
    requestData = {"request": {"branch":"master"}};
    r = requests.post("https://api.travis-ci.org/repo/amalgam8%2Ftesting/requests",
                      headers = {
                          "content-type": "application/json",
                          "accept" : "application/json",
                          "Travis-API-Version": "3",
                          "Authorization" : "token <TRAVIS_TOKEN_OBTAINED_FROM_TRAVIS_CLI>"
                      },
                      data = json.dumps(requestData))
    if r.status_code >=200 and r.status_code < 300:
        print(r.json())
    else:
        print 'request failed'
        r.raise_for_status()
    return {'status': r.status_code}

if __name__ == "__main__":
    main({})
