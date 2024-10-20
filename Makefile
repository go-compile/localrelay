# VERSION=$(shell git describe --tags --abbrev=0)
VERSION=v2.0.0-rc3
COMMIT=$(shell git rev-parse --short HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

build:
	goreleaser release --snapshot --clean --skip=publish,sign

release:
	goreleaser release --clean --skip=publish

wix:
	cp ./scripts/wix/localrelay.template.wxs ./scripts/wix/localrelay.wxs
	sed -i -E 's/LR_VERSION/${VERSION}/g' ./scripts/wix/localrelay.wxs
	wix build ./scripts/wix/localrelay.wxs

publish:
	goreleaser release --clean

install:
	cd ./cmd/localrelay/ && go install -v -ldflags="-s -w -X main.VERSION=${VERSION} -X main.BRANCH=${BRANCH} -X main.COMMIT=${COMMIT}"

dev-install:
	cd ./cmd/localrelay/ && go build -v -ldflags="-s -w -X main.VERSION=${VERSION} -X main.BRANCH=${BRANCH} -X main.COMMIT=${COMMIT}"
	sudo chown root:root ./cmd/localrelay/localrelay
	sudo chmod 755 ./cmd/localrelay/localrelay
	sudo mv ./cmd/localrelay/localrelay /usr/bin/
	sudo localrelay restart

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
	rm ./scripts/wix/localrelay.wxs
	rm ./scripts/wix/localrelay.msi