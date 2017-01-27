APP_NAME 	:= $(shell basename `pwd`)
IMAGE_NAME 	:= amalgam8/a8-testing-bookinfo-$(APP_NAME)

BUILDFLAGS  :=
LDFLAGS 	:= -extldflags -static

# Integration test runs within docker containers only
GOOS        := linux
GOARCH      := amd64

build:
	@go build $(BUILDFLAGS) -ldflags '$(LDFLAGS)' -o $(APP_NAME)
	
dockerize:
	@docker build -t $(IMAGE_NAME)-v1:latest .
	@docker build -t $(IMAGE_NAME)-sidecar-v1:latest -f Dockerfile.sidecar .

clean:
	@rm $(APP_NAME)
