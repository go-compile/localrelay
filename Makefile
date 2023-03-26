VERSION="Unkown"

build:
	goreleaser release --snapshot --clean --skip-sign --skip-publish

publish:
	goreleaser release --clean

install:
	go install ./cmd/localrelay

install-deps:
	# Install developer dependencies

	# install build system
	go install github.com/goreleaser/goreleaser@latest

docker:
	docker build . --tag localrelay --build-arg VERSION=${VERSION}

docker-push:
	# TODO: auto insert tag version
	docker buildx build --platform linux/amd64,linux/arm/v7,linux/arm64 . --tag gocompile/localrelay:latest --build-arg VERSION=${VERSION} --push
	docker buildx build --platform linux/amd64,linux/arm/v7,linux/arm64 . --tag gocompile/localrelay:${VERSION} --build-arg VERSION=${VERSION} --push

clean:
	rm -rf ./dist/
	go clean