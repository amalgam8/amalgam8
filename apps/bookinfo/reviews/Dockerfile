FROM jfloff/alpine-python:2.7-onbuild

RUN mkdir -p /opt/microservices
COPY . /opt/microservices/
EXPOSE 9080
WORKDIR /opt/microservices
ARG enable_ratings
ARG star_color
ENV ENABLE_RATINGS ${enable_ratings:-false}
ENV STAR_COLOR ${star_color:-black}
CMD python reviews.py 9080 http://localhost:6379
