# build environment ###########################################
FROM golang:1.20.1-alpine@sha256:87d0a3309b34e2ca732efd69fb899d3c420d3382370fd6e7e6d2cb5c930f27f9 AS build-env

WORKDIR /app

ENV SEC_AUDIT_LOG /dev/stdout

# entrypoint
RUN apk add --no-cache entr
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

# dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# server
COPY main.go .
COPY mutatingwebhook.go .
RUN go install .

# production image ############################################
FROM alpine:3.17.2@sha256:69665d02cb32192e52e07644d76bc6f25abeb5410edc1c7a81a10ba3f0efb90a AS prod
RUN apk add --no-cache ca-certificates

COPY --from=build-env /go/bin/ingress-waf-admission-controller /bin/admission-controller

ENTRYPOINT ["admission-controller"]
