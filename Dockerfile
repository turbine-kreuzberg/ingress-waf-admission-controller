# build environment ###########################################
FROM golang:1.20.2-alpine@sha256:6feb435834a5f2670ba73ea45dc9bdb8f57c81210d386d921bdd4893f5198f80 AS build-env

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
FROM alpine:3.17.3@sha256:124c7d2707904eea7431fffe91522a01e5a861a624ee31d03372cc1d138a3126 AS prod
RUN apk add --no-cache ca-certificates

COPY --from=build-env /go/bin/ingress-waf-admission-controller /bin/admission-controller

ENTRYPOINT ["admission-controller"]
