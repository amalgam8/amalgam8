APP_NAME 	:= $(shell basename `pwd`)
IMAGE_NAME 	:= amalgam8/a8-testing-bookinfo-$(APP_NAME)

BUILDFLAGS  := -i
LDFLAGS 	:= -linkmode external -extldflags -static

# Integration test runs within docker containers only
GOOS        := linux
GOARCH      := amd64

build:
	@go build $(BUILDFLAGS) -ldflags '$(LDFLAGS)' -o $(APP_NAME)
	
dockerize: dockerize-v1 dockerize-v2 dockerize-v3

dockerize-v1:
	@docker build -t $(IMAGE_NAME):v1 .
	@docker build -t $(IMAGE_NAME)-sidecar:v1 -f Dockerfile.sidecar .
	
dockerize-v2:
	@docker build -t $(IMAGE_NAME):v2 --build-arg enable_ratings=true .
	@docker build -t $(IMAGE_NAME)-sidecar:v2 -f Dockerfile.sidecar --build-arg enable_ratings=true .
	
dockerize-v3:
	@docker build -t $(IMAGE_NAME):v3 --build-arg enable_ratings=true --build-arg star_color=red .
	@docker build -t $(IMAGE_NAME)-sidecar:v3 -f Dockerfile.sidecar --build-arg enable_ratings=true --build-arg star_color=red .
	
clean:
	@rm $(APP_NAME)
