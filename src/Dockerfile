FROM --platform=linux/amd64 golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/pinger 
COPY . .
RUN go get -d -v
RUN go build -o /go/bin/pinger
FROM --platform=linux/amd64 alpine
# Copy our static executable.
COPY --from=builder /go/bin/pinger /go/bin/pinger
COPY --from=builder /bin/ping /bin/ping
RUN chmod +x /go/bin/pinger
ENTRYPOINT ["/go/bin/pinger"]