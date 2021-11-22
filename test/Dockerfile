FROM php:8.0.12-cli-alpine

RUN apk add --update --no-cache composer jq bash tree
RUN composer global require repman-io/composer-plugin

ENV SPYKER_TAG="202108.0"
ADD https://github.com/spryker-shop/b2b-demo-shop/raw/${SPYKER_TAG}/composer.json composer.json
ADD https://github.com/spryker-shop/b2b-demo-shop/raw/${SPYKER_TAG}/composer.lock composer.lock

RUN jq '. += {"config":{"secure-http": false},"extra": {"repman": {"url": "http://php-package-cache:8080/"}}}' composer.json \
    > composer.json.tmp && \
    mv composer.json.tmp composer.json

COPY test.sh .
