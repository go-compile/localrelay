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

clean:
	rm -rf ./dist/
	go clean