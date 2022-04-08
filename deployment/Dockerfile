# build environment ###########################################
FROM golang:1.18.0-alpine@sha256:a2ca4f4c0828b1b426a3153b068bf32a21868911c57a9fc4dccdc5fbb6553b35 AS build-env

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
FROM alpine:3.15.4@sha256:315a3eab8ebf3bbcb931e34d13684b1e53186b8ec342c64383ce5c64890771ab AS prod
RUN apk add --no-cache ca-certificates

# tracing config
ENV JAEGER_SERVICE_NAME PHP-Package-Cache
ENV JAEGER_SAMPLER_TYPE const
ENV JAEGER_SAMPLER_PARAM 1.0
ENV JAEGER_REPORTER_LOG_SPANS true

# executable
COPY --from=build-env /go/bin/php-package-cache /bin/php-package-cache
ENTRYPOINT ["php-package-cache"]
