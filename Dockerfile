FROM golang:1.16-alpine as build
RUN apk --no-cache add make gcc libc-dev git openssl \
    && rm  -rf /tmp/* /var/cache/apk/*

ENV NAME=app
WORKDIR /var/app/golang
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make build \
    && make ssl

FROM alpine:3
ENV TZ=Asia/Tokyo
COPY --from=build /var/app/golang/bin/app /app/app
COPY --from=build /var/app/golang/ssl/ /app/ssl/

WORKDIR /app
RUN chmod +x /app/app \
    && apk --no-cache add tzdata \
    && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
    && echo "Asia/Tokyo" >  /etc/timezone \
    && rm  -rf /tmp/* /var/cache/apk/*
ENV DEBUG=true
ENV SAML_CONFIG=saml.yaml
ENV HTTP_PORT=8080
EXPOSE 8080

CMD [ "./app" ]