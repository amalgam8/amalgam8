#!/usr/bin/python
from flask import Flask, request, render_template
import simplejson as json
import requests
import sys
from json2html import *
from datetime import datetime

app = Flask(__name__)
from flask_bootstrap import Bootstrap
Bootstrap(app)
proxyurl=None

details = {
    "name" : "details",
    "children" : []
}

reviews = {
    "name" : "reviews",
    "children" : []
}

productpage = {
    "name" : "productpage",
    "children" : [details, reviews]
}

service_dict = {
    "productpage" : productpage,
    "details" : details,
    "reviews" : reviews,
}

def getGremlinHeader(request):
    gremlinHeader = request.headers.get('X-Gremlin-ID')

    headers = {}
    if gremlinHeader is not None:
        headers = {'X-Gremlin-ID': gremlinHeader}
    return headers

@app.route('/')
@app.route('/index.html')
def index():
    """ Display productpage with normal user and test user buttons"""
    global productpage

    table = json2html.convert(json = json.dumps(productpage),
                              table_attributes="class=\"table table-condensed table-bordered table-hover\"")

    return render_template('index.html', serviceTable=table)


@app.route('/productpage')
def front():
    headers = getGremlinHeader(request)

    bookdetails = getDetails(headers)
    bookreviews = getReviews(headers)
    return render_template('productpage.html', details=bookdetails, reviews=bookreviews)

def getReviews(headers):
    global proxyurl
    ##timeout is set to 10 milliseconds
    try:
        res = requests.get(proxyurl+"/"+reviews['name'], headers=headers)#, timeout=0.010)
    except:
        res = None

    if res and res.status_code == 200:
        return res.text
    else:
        return """<h3>Sorry, product reviews are currently unavailable for this book.</h3>"""


def getDetails(headers):
    global proxyurl
    try:
        res = requests.get(proxyurl+"/"+details['name'], headers=headers)#, timeout=0.010)
    except:
        res = None

    if res and res.status_code == 200:
        return res.text
    else:
        return """<h3>Sorry, product details are currently unavailable for this book.</h3>"""

if __name__ == '__main__':
    # To run the server, type-in $ python server.py
    if len(sys.argv) < 2:
        print "usage: %s port proxyurl" % (sys.argv[0])
        sys.exit(-1)

    p = int(sys.argv[1])
    proxyurl = sys.argv[2]
    app.run(port=p)#host='0.0.0.0', port=p)
