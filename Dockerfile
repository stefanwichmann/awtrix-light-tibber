FROM alpine:latest

WORKDIR /opt/awtrix-light-tibber

RUN apk --no-cache add ca-certificates tzdata && update-ca-certificates
COPY awtrix-light-tibber /opt/awtrix-light-tibber/

ENTRYPOINT /opt/awtrix-light-tibber/awtrix-light-tibber