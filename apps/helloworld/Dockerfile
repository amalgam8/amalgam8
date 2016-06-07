FROM phusion/baseimage

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -qq && \
   apt-get install -y -q --no-install-recommends \
   python2.7 python-pip build-essential python-dev
   
RUN pip install Flask Flask-Cache pymongo

# application source code including static files, templates, etc
ADD app.py /opt/

EXPOSE 5000

ENTRYPOINT ["python", "-u", "/opt/app.py"]
