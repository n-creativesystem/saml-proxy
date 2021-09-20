VERSION ?= 
REVISION ?=
SRCS	:= $(shell find . -type d -name archive -prune -o -type f -name '*.go')
LDFLAGS	:= -ldflags="-s -w -X \"github.com/n-creativesystem/saml-proxy/version.Version=$(VERSION)\" -X \"github.com/n-creativesystem/saml-proxy/version.Revision=$(REVISION)\" -extldflags \"-static\""

build/static: $(SRCS)
	CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o bin/$(NAME)

build: $(SRCS)
	go build $(LDFLAGS) -o bin/$(NAME)
build/docker:
	docker build -t ${IMAGE_NAME} .

.PHONY: ssl
ssl:
	cd ssl && go run ./ ca && go run ./ server
