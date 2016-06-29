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


from flask import Flask, request, jsonify
import logging
import sys
from json2html import *

app = Flask(__name__)

@app.route('/ratings')
def bookRatings():
    return "Hello this is ratings v1"

class Writer(object):

    def __init__(self, filename):
        self.file = open(filename,'w')

    def write(self, data):
        self.file.write(data)
        self.file.flush()


proxyurl=None
if __name__ == '__main__':
    if len(sys.argv) < 2:
        print "usage: %s port proxyurl" % (sys.argv[0])
        sys.exit(-1)

    p = int(sys.argv[1])
    proxyurl = sys.argv[2]
    sys.stderr = Writer('stderr.log')
    sys.stdout = Writer('stdout.log')
    logging.basicConfig(filename='microservice.log',filemode='w',level=logging.DEBUG)
    app.run(host='0.0.0.0', port=p, debug = True, threaded=True)

