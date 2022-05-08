FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
COPY VERSION ./
COPY .ddns-aws.yaml ./.ddns-aws.yaml
RUN mkdir -p /certificates

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /ddns-aws




FROM scratch
WORKDIR /
VOLUME /etc
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /ddns-aws /ddns-aws
COPY --from=builder /app/.ddns-aws.yaml /etc/.ddns-aws.yaml
COPY --from=builder /certificates /certificates

EXPOSE 8125

ENTRYPOINT ["/ddns-aws","server","run"]
