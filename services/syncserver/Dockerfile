FROM golang:1.26-alpine AS builder

WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/syncserver ./services/syncserver/

FROM alpine:3.22

RUN apk --no-cache add ca-certificates \
    && addgroup -g 10001 -S zeropass \
    && adduser -u 10001 -S -D -H -h /var/lib/zeropass -G zeropass zeropass \
    && mkdir -p /var/lib/zeropass \
    && chown -R 10001:10001 /var/lib/zeropass

COPY --from=builder /out/syncserver /usr/local/bin/syncserver

WORKDIR /var/lib/zeropass

USER 10001:10001

EXPOSE 8443

ENTRYPOINT ["syncserver"]