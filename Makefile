build:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-windows-x64.exe ./cmd/localrelay
	GOOS=windows GOARCH=386 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-windows.exe ./cmd/localrelay
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-windows-arm64.exe ./cmd/localrelay
	
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-darwin ./cmd/localrelay
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-darwin-arm64 ./cmd/localrelay
	
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux-64 ./cmd/localrelay
	GOOS=linux GOARCH=386 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux ./cmd/localrelay
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux-arm64 ./cmd/localrelay

	GOOS=freebsd GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-freebsd-64 ./cmd/localrelay
	GOOS=freebsd GOARCH=386 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-freebsd ./cmd/localrelay
	GOOS=freebsd GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-freebsd-arm64 ./cmd/localrelay
	
	GOOS=openbsd GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-openbsd-64 ./cmd/localrelay
	GOOS=openbsd GOARCH=386 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-openbsd ./cmd/localrelay
	GOOS=openbsd GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-openbsd-arm64 ./cmd/localrelay

clean:
	rm -rf ./bin/