FROM jfloff/alpine-python:2.7-onbuild

RUN mkdir -p /opt/microservices
COPY . /opt/microservices/
EXPOSE 9080
WORKDIR /opt/microservices
CMD python details.py 9080 http://localhost:6379
