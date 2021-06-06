FROM golang:1.16 AS builder

ENV GOOS=linux GARCH=amd64 CGO_ENABLED=0

COPY ./ /url_shortener
WORKDIR /url_shortener

RUN go mod download
RUN go build -o server -a -installsuffix cgo server/server.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /url_shortener/server .
COPY ./docker/start.sh .
RUN chmod +x ./start.sh

CMD ["./start.sh"]