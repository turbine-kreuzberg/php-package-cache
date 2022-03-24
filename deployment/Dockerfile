# build environment ###########################################
FROM golang:1.18.0-alpine@sha256:9ccb0ed869157f3b1630f5dda3422b5974defa9dd82c7375ca68dc3a9cbf8fae AS build-env

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
FROM alpine:3.15.2@sha256:ceeae2849a425ef1a7e591d8288f1a58cdf1f4e8d9da7510e29ea829e61cf512 AS prod
RUN apk add --no-cache ca-certificates

# tracing config
ENV JAEGER_SERVICE_NAME PHP-Package-Cache
ENV JAEGER_SAMPLER_TYPE const
ENV JAEGER_SAMPLER_PARAM 1.0
ENV JAEGER_REPORTER_LOG_SPANS true

# executable
COPY --from=build-env /go/bin/php-package-cache /bin/php-package-cache
ENTRYPOINT ["php-package-cache"]
