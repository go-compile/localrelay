version := 1.2.0
revision := 1
# Install in ubuntu/deb: sudo apt install gcc-aarch64-linux-gnu binutils-aarch64-linux-gnu
# arch64_cc is only required if cgo_enabled=1
arch64_cc := aarch64-linux-gnu-gcc

build:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-windows-x64.exe ./cmd/localrelay
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux-64 ./cmd/localrelay
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-darwin ./cmd/localrelay


# Compiling for Windows requires either a Windows system or MSYS2
# Cross compiling may required a Docker Container: https://dh1tw.de/2019/12/cross-compiling-golang-cgo-projects/
build-win:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-windows-x64.exe ./cmd/localrelay
	GOOS=windows GOARCH=386 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-windows.exe ./cmd/localrelay

build-win-arm64:
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-windows-arm64.exe ./cmd/localrelay

build-darwin:	
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-darwin ./cmd/localrelay
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-darwin-arm64 ./cmd/localrelay
	
build-linux:
	echo "[BUILDING] AMD64"
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux-64 ./cmd/localrelay
	echo "[BUILDING] 386"
	GOOS=linux GOARCH=386 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux ./cmd/localrelay
	echo "[BUILDING] arm64"
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux-arm64 ./cmd/localrelay

build-deb:
# Build .deb packages

	mkdir -p ./packages/deb_amd64/usr/bin
	mkdir -p ./packages/deb_i386/usr/bin
	mkdir -p ./packages/deb_amd64/usr/bin

	cp ./bin/localrelay-linux-64 ./packages/localrelay 
	cp -r ./packages/deb_amd64/ ./packages/localrelay_$(version)_$(revision)_amd64
	mv ./packages/localrelay ./packages/localrelay_$(version)_$(revision)_amd64/usr/bin
	dpkg-deb --build --root-owner-group ./packages/localrelay_$(version)_$(revision)_amd64
	mv ./packages/localrelay_$(version)_$(revision)_amd64.deb ./bin/localrelay_$(version)_$(revision)_amd64.deb
	rm -rf ./packages/localrelay_$(version)_$(revision)_amd64	

	cp ./bin/localrelay-linux ./packages/localrelay
	cp -r ./packages/deb_i386/ ./packages/localrelay_$(version)_$(revision)_i386
	mv ./packages/localrelay ./packages/localrelay_$(version)_$(revision)_i386/usr/bin
	dpkg-deb --build --root-owner-group ./packages/localrelay_$(version)_$(revision)_i386
	mv ./packages/localrelay_$(version)_$(revision)_i386.deb ./bin/localrelay_$(version)_$(revision)_i386.deb
	rm -rf ./packages/localrelay_$(version)_$(revision)_i386

	
	cp ./bin/localrelay-linux-arm64 ./packages/localrelay
	cp -r ./packages/deb_arm64/ ./packages/localrelay_$(version)_$(revision)_arm64
	mv ./packages/localrelay  ./packages/localrelay_$(version)_$(revision)_arm64/usr/bin
	dpkg-deb --build --root-owner-group ./packages/localrelay_$(version)_$(revision)_arm64
	mv ./packages/localrelay_$(version)_$(revision)_arm64.deb ./bin/localrelay_$(version)_$(revision)_arm64.deb
	rm -rf ./packages/localrelay_$(version)_$(revision)_arm64

build-freebsd:
	GOOS=freebsd GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-freebsd-64 ./cmd/localrelay
	GOOS=freebsd GOARCH=386 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-freebsd ./cmd/localrelay
	GOOS=freebsd GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-freebsd-arm64 ./cmd/localrelay

build-openbsd:
	GOOS=openbsd GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-openbsd-64 ./cmd/localrelay
	GOOS=openbsd GOARCH=386 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-openbsd ./cmd/localrelay
	GOOS=openbsd GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-openbsd-arm64 ./cmd/localrelay

# Cross compile from windows
cross-compile-win:
	make build-win
	# build-win-arm64
	make build-freebsd
	make build-openbsd

	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux-64 ./cmd/localrelay
	GOOS=linux GOARCH=386 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux ./cmd/localrelay
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux-arm64 ./cmd/localrelay

cross-compile-linux:
	build-win

	make build-freebsd
	make build-openbsd

	make build-linux
	make build-deb
	
cross-compile-cgo:
# To cross compile from windows remove the whole CC variable
	CC=$(arch64_cc) GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o ./bin/localrelay-linux-arm64 ./cmd/localrelay
# To cross compile to arch32 you may want to follow the above step but replace CC with the compiler provided in this guide:
# https://jensd.be/1126/linux/cross-compiling-for-arm-or-aarch64-on-debian-or-ubuntu

clean:
	rm -rf ./bin/