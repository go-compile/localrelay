VERSION=$(shell git describe --tags --abbrev=0)
COMMIT=$(shell git rev-parse --short HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

build:
	goreleaser release --snapshot --clean --skip-sign --skip-publish
	cd ./scripts/ && wix build localrelay.wxs

publish:
	goreleaser release --clean

install:
	cd ./cmd/localrelay/ && go install -v -ldflags="-s -w -X main.VERSION=${VERSION} -X main.BRANCH=${BRANCH} -X main.COMMIT=${COMMIT}"

install-deps:
	# Install developer dependencies

	# install build system
	go install github.com/goreleaser/goreleaser@latest

docker:
	docker build . --tag localrelay --build-arg VERSION=${VERSION} --build-arg COMMIT=${COMMIT} --build-arg BRANCH=${BRANCH}

docker-push:
	docker buildx build --platform linux/amd64,linux/arm/v7,linux/arm64 . --tag gocompile/localrelay:latest --build-arg VERSION=${VERSION} --push
	docker buildx build --platform linux/amd64,linux/arm/v7,linux/arm64 . --tag gocompile/localrelay:${VERSION} --build-arg VERSION=${VERSION} --push

clean:
	rm -rf ./dist/
	go clean