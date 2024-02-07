# build environment ###########################################
FROM golang:1.21.7-alpine@sha256:f63ad366ae27636efbab3565c26a267dfbfbf3cce27bfa4857f3b7e8ee04208f AS build-env

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
FROM alpine:3.19.1@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b AS prod
RUN apk add --no-cache ca-certificates

COPY --from=build-env /go/bin/ingress-waf-admission-controller /bin/admission-controller

ENTRYPOINT ["admission-controller"]
