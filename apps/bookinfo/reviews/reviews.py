#!/usr/bin/python
from flask import Flask, request
import simplejson as json
import requests
import sys, os
from json2html import *

app = Flask(__name__)

ratings_service = {
    "name" : "ratings",
    "endpoint" : "ratings",
    "children" : []
}

reviews_resp="""
<blockquote>
<p>
An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!
</p> <small>Reviewer1 <cite>Affiliation1</cite></small>
{}
</blockquote>
<blockquote>
<p>
Absolutely fun and entertaining. The play lacks thematic depth when compared 
to other plays by Shakespeare.
</p> <small>Reviewer2 <cite>Affiliation2</cite></small>
{}
</blockquote>
"""

ratings_enabled = os.getenv('ENABLE_RATINGS') == str.lower("true")
star_color = os.getenv('STAR_COLOR') or 'black'

def getForwardHeaders(request):
    headers = {}
    
    user_cookie = request.cookies.get("user")
    if user_cookie:
        headers['Cookie'] = 'user=' + user_cookie
        
    reqTrackingHeader = request.headers.get('X-Request-ID')
    if reqTrackingHeader is not None:
        headers['X-Request-ID'] = reqTrackingHeader

    return headers

@app.route('/reviews')
def bookReviews():
    global reviews_resp
    global ratings_enabled
    headers = getForwardHeaders(request)
    r1 = ""
    r2 = ""
    if ratings_enabled:
        ratings = getRatings(headers)
        if "Reviewer1" in ratings:
            r1 = ratings["Reviewer1"]
        if "Reviewer2" in ratings:
            r2 = ratings["Reviewer2"]
        #print ratings
    resp = reviews_resp.format(r1, r2)
    return resp

def getRatings(headers):
    global proxyurl
    global ratings_service
    global star_color    
    try:
        timeout = 10.0 if star_color == 'black' else 2.5
        res = requests.get(proxyurl+"/"+ratings_service['name']+"/"+ratings_service['endpoint'], headers=headers, timeout=timeout)
    except:
        res = None

    ratings = {}
    try:
        if res and res.status_code == 200:
            ratings = res.json()
            for k in ratings:
                v = ratings[k]
                stars = "<font color=\"" + star_color + "\">"
                for _ in range(v):
                    stars = stars + "<span class=\"glyphicon glyphicon-star\"></span>"
                stars = stars + "</font>"
                if v < 5:
                    for _ in range(5-v):
                        stars = stars + "<span class=\"glyphicon glyphicon-star-empty\"></span>"
                ratings[k] = stars
    except:
        pass
    return ratings

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
    <title>Book reviews service</title>
    <body>
    <p><h2>Hello! This is the book reviews service. My content is</h2></p>
    <div>%s</div>
    <p>Ratings service enabled? %s</p>
    <p>Star color: %s </p>
    </body>
    </html>
    """ % (reviews_resp, ratings_enabled, star_color)
    return top

proxyurl=None
if __name__ == '__main__':
    if len(sys.argv) < 2:
        print "usage: %s port proxyurl" % (sys.argv[0])
        sys.exit(-1)

    p = int(sys.argv[1])
    proxyurl = sys.argv[2]
    app.run(host='0.0.0.0', port=p, debug = True)
