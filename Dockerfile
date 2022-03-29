ARG VERSION="Dev"
ARG GIT_SHA

FROM golang:alpine as builder

LABEL maintainer="luca@lucabernstein.com"

WORKDIR /src
COPY . .

RUN go mod download
RUN go build -buildvcs=false -o app

FROM alpine

ENV GIN_MODE release

ARG VERSION
ENV VERSION=$VERSION
ARG GIT_SHA
ENV GIT_SHA=$GIT_SHA

WORKDIR /

COPY --from=builder /src/app /bin/app

EXPOSE 8081

ENTRYPOINT [ "/bin/app" ]
