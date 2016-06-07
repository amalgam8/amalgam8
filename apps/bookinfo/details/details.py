#!/usr/bin/python
from flask import Flask, request
import simplejson as json
import requests
import sys
from json2html import *

app = Flask(__name__)

details_resp="""
<h4 class="text-center text-primary">Product Details</h4>
<dl>
<dt>Paperback:</dt>200 pages
<dt>Publisher:</dt>O'Reilly Media; 1 edition (March 25, 2015)
<dt>Language:</dt>English
<dt>ISBN-10:</dt>1491914254
<dt>ISBN-13:</dt>978-1491914250
</dl>
"""

@app.route('/details')
def bookDetails():
    global details_resp
    return details_resp

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
    <title>Book details service</title>
    <body>
    <p><h2>Hello! This is the book details service. My content is</h2></p>
    <div>%s</div>
    </body>
    </html>
    """ % (details_resp)
    return top

proxyurl=None
if __name__ == '__main__':
    # To run the server, type-in $ python server.py
    if len(sys.argv) < 2:
        print "usage: %s port proxyurl" % (sys.argv[0])
        sys.exit(-1)

    p = int(sys.argv[1])
    proxyurl = sys.argv[2]
    app.run(host='0.0.0.0', port=p, debug = False)
