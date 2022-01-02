FROM golang:1.17.5-alpine as build-env
RUN go install github.com/m1dugh/crawler/cmd/crawler@master

FROM alpine:3.15.0
COPY --from=build-env /go/bin/crawler /usr/local/bin/crawler
ENTRYPOINT ["crawler"]

