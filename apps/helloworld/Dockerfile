FROM jfloff/alpine-python:2.7-onbuild

RUN mkdir -p /opt/microservices
ADD app.py /opt/microservices/
EXPOSE 5000
WORKDIR /opt/microservices
CMD python app.py
