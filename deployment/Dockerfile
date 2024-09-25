# build environment ###########################################
FROM golang:1.18.1-alpine@sha256:42d35674864fbb577594b60b84ddfba1be52b4d4298c961b46ba95e9fb4712e8 AS build-env

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
FROM alpine:3.20.3@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d AS prod
RUN apk add --no-cache ca-certificates

# tracing config
ENV JAEGER_SERVICE_NAME PHP-Package-Cache
ENV JAEGER_SAMPLER_TYPE const
ENV JAEGER_SAMPLER_PARAM 1.0
ENV JAEGER_REPORTER_LOG_SPANS true

# executable
COPY --from=build-env /go/bin/php-package-cache /bin/php-package-cache
ENTRYPOINT ["php-package-cache"]
