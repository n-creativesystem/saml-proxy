SRCS	:= $(shell find . -type d -name archive -prune -o -type f -name '*.go')
LDFLAGS	:= -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""

build/static: $(SRCS)
	CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o bin/$(NAME)

build: $(SRCS)
	go build -o bin/$(NAME)
build/docker:
	docker build -t ${IMAGE_NAME} .
.PHONY: ssl
ssl:
	@openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout ssl/server.key -out ssl/server.crt -subj "/C=JP/ST=Osaka/L=Osaka/O=NCreativeSystem, Inc./CN=localhost"
