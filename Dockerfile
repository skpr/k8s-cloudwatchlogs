FROM golang:1.12 as build
RUN go get github.com/mitchellh/gox
ADD . /go/src/github.com/skpr/k8s-cloudwatchlogs
WORKDIR /go/src/github.com/skpr/k8s-cloudwatchlogs
RUN make build

FROM alpine:3.9
RUN apk --no-cache add ca-certificates
COPY --from=build /go/src/github.com/skpr/k8s-cloudwatchlogs/bin/k8s-cloudwatchlogs_linux_amd64 /usr/local/bin/k8s-cloudwatchlogs
CMD ["k8s-cloudwatchlogs"]
