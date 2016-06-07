FROM jfloff/alpine-python:2.7-onbuild

RUN mkdir -p /opt/microservices
COPY . /opt/microservices/
EXPOSE 9080
WORKDIR /opt/microservices
CMD python productpage.py 9080 http://localhost:6379 http://192.168.33.33:32000
