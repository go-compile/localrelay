FROM golang:1.18-alpine AS builder

WORKDIR /app/src

COPY ./cmd/localrelay /app/src/

RUN go mod init localrelay-cli
RUN go mod tidy
# Download latest version from master branch
RUN go get github.com/go-compile/localrelay@master

# Build as static binary
RUN CGO_ENABLED=0 go build -v -ldflags="-s -w -extldflags -static" -tags osusergo,netgo -o /app/localrelay

RUN apk add upx
RUN upx --ultra-brute /app/localrelay

RUN rm -rf /app/src

# Create empty dir to be used as mkdir COPY src.
# SCRATCH images do not have mkdir thus we must use "COPY /app/empty /dst"
RUN mkdir /app/empty

FROM scratch

COPY --from=builder /app /usr/bin
COPY --from=builder /app/empty /var/run

WORKDIR /app

CMD ["/usr/bin/localrelay", "start-service-daemon"]
