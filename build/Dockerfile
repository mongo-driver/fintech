FROM golang:1.22 AS builder

ARG SERVICE
WORKDIR /src

COPY go.mod go.sum ./
ENV GOPROXY=https://proxy.golang.org,direct \
    GOSUMDB=sum.golang.org \
    GODEBUG=http2client=0
RUN set -eux; \
    for i in 1 2 3 4 5; do \
      go mod download && break; \
      echo "go mod download failed, retry $i/5"; \
      sleep 5; \
    done

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/service ./services/${SERVICE}/cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/service /service
EXPOSE 8080
ENTRYPOINT ["/service"]
