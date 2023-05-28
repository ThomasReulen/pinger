FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/pinger 
COPY . .
RUN go get -d -v
RUN go build -o /go/bin/pinger
FROM alpine
# Copy our static executable.
COPY --from=builder /go/bin/pinger /go/bin/pinger
COPY --from=builder /bin/ping /bin/ping
ENTRYPOINT ["/go/bin/pinger"]