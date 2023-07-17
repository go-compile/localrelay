FROM golang:1.18-alpine AS builder

ARG VERSION
ARG COMMIT
ARG BRANCH

WORKDIR /app/src

COPY ./cmd/localrelay /app/src/

RUN go mod init localrelay-cli
RUN go mod tidy
# Download latest version from master branch
RUN go get github.com/go-compile/localrelay@master

# Build as static binary
RUN CGO_ENABLED=0 go build -v -ldflags="-s -w -extldflags -static -X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}" -tags osusergo,netgo -o /app/localrelay

# RUN arch=$TARGETPLATFORM; amd64="linux/amd64"
# RUN if [ $arch = $amd64 ]; then apk add upx; fi
# RUN if [ $arch = $amd64 ]; then upx --ultra-brute /app/localrelay; fi

RUN rm -rf /app/src

# Create empty dir to be used as mkdir COPY src.
# SCRATCH images do not have mkdir thus we must use "COPY /app/empty /dst"
RUN mkdir /app/empty

FROM scratch

COPY --from=builder /app /usr/bin
COPY --from=builder /app/empty /var/run

# fix TLS unable to verify CA errors
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \ 
	CMD [ "localrelay", "status" ]

CMD ["/usr/bin/localrelay", "start-service-daemon"]
