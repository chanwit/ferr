VERSION?=$(shell grep 'VERSION' cmd/ferr/main.go | awk '{ print $$4 }' | tr -d '"')
GH_VERSION=1.2.0
COMPOSE_VERSION=1.27.4
FLUX_VERSION=0.2.1
KOMPOSE_VERSION=1.22.0
KIND_VERSION=0.9.0
STERN_VERSION=1.11.0
KUBECTL_VERSION=1.19.3
SKAFFOLD_VERSION=1.16.0
PACK_CLI=0.14.2

all: test build-linux

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test: tidy fmt vet docs
	go test ./... -coverprofile cover.out

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/ferr-linux.bin ./cmd/ferr

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/ferr-darwin.bin ./cmd/ferr

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/ferr-windows.exe ./cmd/ferr

install:
	go install cmd/ferr

.PHONY: docs
docs:
	rm docs/cmd/*
	mkdir -p ./docs/cmd && go run ./cmd/ferr/ docgen

install-dev: build-linux
	sudo cp ./bin/ferr-linux.bin /usr/local/bin/ferr

############ Linux ############

.ONESHELL: download-bins-linux download-bins-darwin
download-bins-linux:
	mkdir -p build/dist-linux/bin
	cd build
	# gh cli 1.2.0
	wget -O- https://github.com/cli/cli/releases/download/v$(GH_VERSION)/gh_$(GH_VERSION)_linux_amd64.tar.gz \
		| tar xzf - --strip-components=2 gh_$(GH_VERSION)_linux_amd64/bin/gh && mv gh dist-linux/bin
	# docker-compose 1.27.4
	wget https://github.com/docker/compose/releases/download/$(COMPOSE_VERSION)/docker-compose-Linux-x86_64 \
		&& mv docker-compose-Linux-x86_64 dist-linux/bin/docker-compose \
		&& chmod +x dist-linux/bin/docker-compose
	# flux 0.2.1
	wget -O- https://github.com/fluxcd/flux2/releases/download/v$(FLUX_VERSION)/flux_$(FLUX_VERSION)_linux_amd64.tar.gz \
		| tar xzf - flux && mv flux dist-linux/bin
	# kompose 1.22.0
	wget https://github.com/kubernetes/kompose/releases/download/v$(KOMPOSE_VERSION)/kompose-linux-amd64 \
		&& mv kompose-linux-amd64 dist-linux/bin/kompose \
		&& chmod +x dist-linux/bin/kompose
	# kind 0.9.0
	wget https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-linux-amd64 \
		&& mv kind-linux-amd64 dist-linux/bin/kind \
		&& chmod +x dist-linux/bin/kind	
	# kubectl
	wget https://storage.googleapis.com/kubernetes-release/release/v$(KUBECTL_VERSION)/bin/linux/amd64/kubectl \
		&& mv kubectl dist-linux/bin/ \
		&& chmod +x dist-linux/bin/kubectl

	# stern 1.11.0
	wget https://github.com/wercker/stern/releases/download/$(STERN_VERSION)/stern_linux_amd64 \
		&& mv stern_linux_amd64 dist-linux/bin/stern \
		&& chmod +x dist-linux/bin/stern

	wget https://github.com/chanwit/warp/releases/download/v0.3.0-musl/warp-packer \
		&& mv warp-packer warp-packer-musl \
		&& chmod +x warp-packer-musl

pack-linux: build-linux
	# pack
	cp ./bin/ferr-linux.bin build/dist-linux/bin/ferr.bin

	cp scripts/linux-launch.sh build/dist-linux/launch.sh
	chmod +x build/dist-linux/launch.sh

	cd build \
	&& ./warp-packer-musl --arch linux-x64 --input_dir dist-linux --exec launch.sh --output ferr-linux-amd64

install-dist-linux: pack-linux
	sudo cp ./build/ferr-linux-amd64 /usr/local/bin/ferr

############ Darwin ############

download-bins-darwin:
	mkdir -p build/dist-darwin/bin
	cd build
	# gh cli 1.2.0
	wget -O- https://github.com/cli/cli/releases/download/v$(GH_VERSION)/gh_$(GH_VERSION)_macOS_amd64.tar.gz \
		| tar xzf - --strip-components=2 gh_$(GH_VERSION)_macOS_amd64/bin/gh && mv gh dist-darwin/bin
	# docker-compose 1.27.4
	wget https://github.com/docker/compose/releases/download/$(COMPOSE_VERSION)/docker-compose-Darwin-x86_64 \
		&& mv docker-compose-Darwin-x86_64 dist-darwin/bin/docker-compose \
		&& chmod +x dist-darwin/bin/docker-compose
	# flux 0.2.1
	wget -O- https://github.com/fluxcd/flux2/releases/download/v$(FLUX_VERSION)/flux_$(FLUX_VERSION)_darwin_amd64.tar.gz \
		| tar xzf - flux && mv flux dist-darwin/bin
	# kompose 1.22.0
	wget https://github.com/kubernetes/kompose/releases/download/v$(KOMPOSE_VERSION)/kompose-darwin-amd64 \
		&& mv kompose-darwin-amd64 dist-darwin/bin/kompose \
		&& chmod +x dist-darwin/bin/kompose
	# kind 0.9.0
	wget https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-darwin-amd64 \
		&& mv kind-darwin-amd64 dist-darwin/bin/kind \
		&& chmod +x dist-darwin/bin/kind
	# kubectl
	wget https://storage.googleapis.com/kubernetes-release/release/v$(KUBECTL_VERSION)/bin/darwin/amd64/kubectl \
		&& mv kubectl dist-darwin/bin/ \
		&& chmod +x dist-darwin/bin/kubectl
	# stern 1.11.0
	wget https://github.com/wercker/stern/releases/download/$(STERN_VERSION)/stern_darwin_amd64 \
		&& mv stern_darwin_amd64 dist-darwin/bin/stern \
		&& chmod +x dist-darwin/bin/stern

	wget https://github.com/dgiagio/warp/releases/download/v0.3.0/linux-x64.warp-packer \
		&& mv linux-x64.warp-packer warp-packer \
		&& chmod +x warp-packer

pack-darwin: build-darwin
	# pack
	cp ./bin/ferr-darwin.bin build/dist-darwin/bin/ferr.bin

	cp scripts/darwin-launch.sh build/dist-darwin/launch.sh
	chmod +x build/dist-darwin/launch.sh

	cd build \
	&& ./warp-packer --arch macos-x64 --input_dir dist-darwin --exec launch.sh --output ferr-darwin-amd64

############ Windows ############
download-bins-windows:
	mkdir -p build/dist-windows/bin
	cd build
	# gh cli 1.2.0
	wget https://github.com/cli/cli/releases/download/v1.2.0/gh_1.2.0_windows_amd64.zip \
		&& unzip -p gh_1.2.0_windows_amd64.zip bin/gh.exe > dist-windows/bin/gh.exe
	# docker-compose 1.27.4
	wget https://github.com/docker/compose/releases/download/$(COMPOSE_VERSION)/docker-compose-Windows-x86_64.exe \
		&& mv docker-compose-Windows-x86_64.exe dist-windows/bin/docker-compose.exe
	# flux 0.2.1
	wget https://github.com/fluxcd/flux2/releases/download/v$(FLUX_VERSION)/flux_$(FLUX_VERSION)_windows_amd64.zip \
		&& unzip -p flux_$(FLUX_VERSION)_windows_amd64.zip flux.exe > dist-windows/bin/flux.exe
	# kompose 1.22.0
	wget https://github.com/kubernetes/kompose/releases/download/v$(KOMPOSE_VERSION)/kompose-windows-amd64.exe \
		&& mv kompose-windows-amd64.exe dist-windows/bin/kompose.exe
	# kind 0.9.0
	wget https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-windows-amd64 \
		&& mv kind-windows-amd64 dist-windows/bin/kind.exe
	# kubectl
	wget https://storage.googleapis.com/kubernetes-release/release/v$(KUBECTL_VERSION)/bin/windows/amd64/kubectl.exe \
		&& mv kubectl.exe dist-windows/bin/
	# stern 1.11.0
	wget https://github.com/wercker/stern/releases/download/$(STERN_VERSION)/stern_windows_amd64.exe \
		&& mv stern_windows_amd64.exe dist-windows/bin/stern.exe

	wget https://github.com/dgiagio/warp/releases/download/v0.3.0/linux-x64.warp-packer \
		&& mv linux-x64.warp-packer warp-packer \
		&& chmod +x warp-packer

pack-windows: build-windows
	# pack
	cp ./bin/ferr-windows.exe build/dist-windows/bin/ferr.bin.exe

	cp scripts/windows-launch.cmd build/dist-windows/launch.cmd

	cd build \
	&& ./warp-packer --arch windows-x64 --input_dir dist-windows --exec launch.cmd --output ferr-windows-amd64.exe

download: download-bins-linux download-bins-darwin download-bins-windows

dist: pack-linux pack-darwin pack-windows
