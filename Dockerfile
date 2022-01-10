FROM golang:1.17.5-alpine as build-env
RUN apk add build-base
RUN go install -v github.com/m1dugh/crawler/cmd/crawler@master

FROM alpine:3.15.0
COPY --from=build-env /go/bin/crawler /usr/local/bin/crawler
ENV GOCRAWLER_ROOT=/gocrawler
VOLUME ${GOCRAWLER_ROOT} 
ENTRYPOINT ["crawler"]

