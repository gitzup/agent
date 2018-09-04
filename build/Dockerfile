FROM golang:1.9.7 AS build
WORKDIR /go/src/github.com/gitzup/agent/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY ./Gopkg.lock ./Gopkg.toml ./
RUN dep ensure -v -vendor-only
COPY ./api ./api/
COPY ./internal ./internal/
COPY ./pkg ./pkg/
COPY ./main.go ./Makefile ./
RUN make

########################################################################################################################
# NOTES:
#   - installing "ca-certificates" & calling "update-ca-certificates" so Pub/Sub won't hang when connecting via 443/SSL
########################################################################################################################

FROM alpine:3.8
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build /go/src/github.com/gitzup/agent/app .
ENTRYPOINT ["/app/app"]
