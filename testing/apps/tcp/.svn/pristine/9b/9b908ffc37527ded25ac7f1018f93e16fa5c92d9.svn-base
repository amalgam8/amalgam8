APP_NAME 	:= $(shell basename `pwd`)
IMAGE_NAME 	:= amalgam8/a8-$(APP_NAME)

BUILDFLAGS  :=
LDFLAGS 	:= -linkmode external -extldflags -static

# Integration test runs within docker containers only
GOOS        := linux
GOARCH      := amd64

build:
	@go build $(BUILDFLAGS) -ldflags '$(LDFLAGS)' -o $(APP_NAME)
	
dockerize:
	@docker build -t $(IMAGE_NAME):latest -f Dockerfile.sidecar .
	
clean:
	@rm $(APP_NAME)
