build:
	goreleaser release --snapshot --clean --skip-sign --skip-publish

publish:
	goreleaser release --clean

clean:
	rm -rf ./dist/
	go clean