FROM golang:1.11 AS build
WORKDIR /go/src/github.com/gitzup/agent/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY ./Gopkg.lock ./Gopkg.toml ./
RUN dep ensure -v -vendor-only
COPY ./main.go ./Makefile ./
COPY ./api ./api/
COPY ./cmd ./cmd/
COPY ./internal ./internal/
COPY ./pkg ./pkg/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make agent

########################################################################################################################
# NOTES:
#   - installing "ca-certificates" & calling "update-ca-certificates" so Pub/Sub won't hang when connecting via 443/SSL
########################################################################################################################

FROM alpine:3.8
WORKDIR /app
COPY --from=build /go/src/github.com/gitzup/agent/agent .
ENTRYPOINT ["/app/agent"]
