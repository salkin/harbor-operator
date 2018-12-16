FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/harbor-operator /usr/local/bin/harbor-operator
