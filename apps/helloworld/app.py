import os
from flask import Flask
app = Flask(__name__)

@app.route("/hello")
def hello():
    service_version = os.environ.get('SERVICE').split(':')
    version = service_version[1] if len(service_version) == 2 else 'UNVERSIONED'
    return "Hello version: %s, container: %s\n" % (version, os.environ.get('HOSTNAME'))

if __name__ == "__main__":
    app.run(host='0.0.0.0')
