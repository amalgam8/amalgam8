APP_NAME 	:= $(shell basename `pwd`)
IMAGE_NAME 	:= amalgam8/a8-testing-bookinfo-$(APP_NAME)
LDFLAGS 	:=

ifndef GOOS
    GOOS := $(shell go env GOHOSTOS)
endif

ifeq ($(GOOS),linux)
	# linker flags to strip symbol tables and debug information
	LDFLAGS     += -s -w
endif

build:
	@go build -o $(APP_NAME) -a -installsuffix nocgo -tags netgo -ldflags '$(LDFLAGS)'
	
dockerize:
	@docker build -t $(IMAGE_NAME):v1 .
	@docker build -t $(IMAGE_NAME)-sidecar:v1 -f Dockerfile.sidecar .

clean:
	@rm $(APP_NAME)
