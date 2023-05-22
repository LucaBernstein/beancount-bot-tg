ARG VERSION="Dev"
ARG GIT_SHA

#
FROM golang:alpine as golang_builder

WORKDIR /src
COPY . .

RUN go mod download
RUN go build -buildvcs=false -o app

#
FROM plugfox/flutter:stable-web AS flutter_builder_web

ADD ./api/ui /api/ui
WORKDIR /api/ui

RUN flutter build web

#
FROM alpine

LABEL maintainer="luca@lucabernstein.com"

ENV GIN_MODE release

ARG VERSION
ENV VERSION=$VERSION
ARG GIT_SHA
ENV GIT_SHA=$GIT_SHA

WORKDIR /dist

COPY --from=flutter_builder_web /api/ui/build/web /dist/api/ui/build/web
COPY --from=golang_builder /src/app /dist/app

EXPOSE 8080

ENTRYPOINT [ "/dist/app" ]
