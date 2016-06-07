#!/usr/bin/python
from flask import Flask, request, render_template, redirect, url_for
import simplejson as json
import requests
import sys
from json2html import *
from datetime import datetime
import requests

app = Flask(__name__)
from flask_bootstrap import Bootstrap
Bootstrap(app)
proxyurl=None

details = {
    "name" : "details",
    "endpoint" : "details",
    "children" : []
}

ratings = {
    "name" : "ratings",
    "endpoint" : "ratings",
    "children" : []
}

reviews = {
    "name" : "reviews",
    "endpoint" : "reviews",
    "children" : [ratings]
}

productpage = {
    "name" : "productpage",
    "endpoint" : "details",
    "children" : [details, reviews]
}

service_dict = {
    "productpage" : productpage,
    "details" : details,
    "reviews" : reviews,
}

def getForwardHeaders(request):
    headers = {}
    
    user_cookie = request.cookies.get("user")
    if user_cookie:
        headers['Cookie'] = 'user=' + user_cookie
        
    reqTrackingHeader = request.headers.get('X-Request-ID')
    if reqTrackingHeader is not None:
        headers['X-Request-ID'] = reqTrackingHeader

    return headers

@app.route('/')
@app.route('/index.html')
def index():
    """ Display productpage with normal user and test user buttons"""
    global productpage

    table = json2html.convert(json = json.dumps(productpage),
                              table_attributes="class=\"table table-condensed table-bordered table-hover\"")

    return render_template('index.html', serviceTable=table)

@app.route('/login', methods=['POST'])
def login():
    user = request.values.get('username')
    response = app.make_response(redirect(publicurl+"/productpage/productpage")) # why does url_for('front') not work?
    response.set_cookie('user', user)
    return response

@app.route('/logout', methods=['GET'])
def logout():
    response = app.make_response(redirect(publicurl+"/productpage/productpage"))
    response.set_cookie('user', '', expires=0)
    return response

@app.route('/productpage')
def front():
    headers = getForwardHeaders(request)
    user = request.cookies.get("user", "")
    bookdetails = getDetails(headers)
    bookreviews = getReviews(headers)
    return render_template('productpage.html', details=bookdetails, reviews=bookreviews, user=user)

def getReviews(headers):
    global proxyurl

    for i in range(2):
        try:
            res = requests.get(proxyurl+"/"+reviews['name']+"/"+reviews['endpoint'], headers=headers, timeout=3.0)
        except:
            res = None

        if res and res.status_code == 200:
            return res.text

    return """<h3>Sorry, product reviews are currently unavailable for this book.</h3>"""


def getDetails(headers):
    global proxyurl
    try:
        res = requests.get(proxyurl+"/"+details['name']+"/"+details['endpoint'], headers=headers, timeout=1.0)
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
    publicurl = sys.argv[3]
    app.run(host='0.0.0.0', port=p)
