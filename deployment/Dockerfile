# build environment ###########################################
FROM golang:1.18.0-alpine@sha256:30a4acd6e557828f01574a74c17c586e91ce1437e9700d169e77569b0fff0a1b AS build-env

WORKDIR /app

# tracing config
ENV JAEGER_SERVICE_NAME PHP-Package-Cache
ENV JAEGER_SAMPLER_TYPE const
ENV JAEGER_SAMPLER_PARAM 1.0
ENV JAEGER_REPORTER_LOG_SPANS true
ENV JAEGER_AGENT_HOST jaeger
ENV JAEGER_AGENT_PORT 6831

# entrypoint
RUN apk add --no-cache entr
COPY ./deployment/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

# dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# server
COPY main.go .
COPY pkg ./pkg
RUN go install .

# production image ############################################
FROM alpine:3.15.1@sha256:d6d0a0eb4d40ef96f2310ead734848b9c819bb97c9d846385c4aca1767186cd4 AS prod
RUN apk add --no-cache ca-certificates

# tracing config
ENV JAEGER_SERVICE_NAME PHP-Package-Cache
ENV JAEGER_SAMPLER_TYPE const
ENV JAEGER_SAMPLER_PARAM 1.0
ENV JAEGER_REPORTER_LOG_SPANS true

# executable
COPY --from=build-env /go/bin/php-package-cache /bin/php-package-cache
ENTRYPOINT ["php-package-cache"]
