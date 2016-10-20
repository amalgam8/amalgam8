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

ratings_resp={"Reviewer1": 5, "Reviewer2" : 4}

@app.route('/ratings')
def bookReviews():
    global ratings_resp
    return jsonify(**ratings_resp)

@app.route('/health')
def health():
    return 'Ratings is healthy', 200

@app.route('/')
def index():
    """ Display frontpage with normal user and test user buttons"""

    top = """
    <html>
    <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <!-- Latest compiled and minified CSS -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">

    <!-- Optional theme -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap-theme.min.css">

    <!-- Latest compiled and minified JavaScript -->
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.4/jquery.min.js"></script>

    <!-- Latest compiled and minified JavaScript -->
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/js/bootstrap.min.js"></script>

    </head>
    <title>Book ratings service</title>
    <body>
    <p><h2>Hello! This is the book ratings service. My content is</h2></p>
    <div>%s</div>
    </body>
    </html>
    """ % (ratings_resp)
    return top


class Writer(object):

    def __init__(self, filename):
        self.file = open(filename,'w')

    def write(self, data):
        self.file.write(data)
        self.file.flush()


if __name__ == '__main__':
    if len(sys.argv) < 1:
        print "usage: %s port" % (sys.argv[0])
        sys.exit(-1)

    p = int(sys.argv[1])
    sys.stderr = Writer('stderr.log')
    sys.stdout = Writer('stdout.log')
    logging.basicConfig(filename='microservice.log',filemode='w',level=logging.DEBUG)
    app.run(host='0.0.0.0', port=p, debug = True, threaded=True)

