# build environment ###########################################
FROM golang:1.20.2-alpine@sha256:576da1aa73f8ffa3dc6a7577b6032bd834aa84f2e1714d3e7e96b06b49f4e177 AS build-env

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
