APP_NAME := $(shell basename `pwd`)
IMAGE_NAME := amalgam8/a8-testing-bookinfo-$(APP_NAME)

build:
	@go build -o $(APP_NAME) -a -installsuffix nocgo -tags netgo -ldflags "-s -w"
	
dockerize:
	@docker build -t $(IMAGE_NAME):v1 .
	@docker build -t $(IMAGE_NAME)-sidecar:v1 -f Dockerfile.sidecar .

clean:
	@rm $(APP_NAME)