GIT_VERSION?=$(shell git describe --tags --always --abbrev=42 --dirty)

build: bin
	go build \
		-o bin/redistop \
		-ldflags "-X github.com/factorysh/redistop/version.version=$(GIT_VERSION)" \
		.

bin:
	mkdir -p bin

docker-build:
	mkdir -p .gocache
	docker run -t \
		-v `pwd`:/src \
		-v `pwd`/.gocache:/.cache \
		-e GOCACHE=/.cache \
		-u `id -u` \
		-w /src \
		bearstech/golang-dev \
		make
	docker run -t \
		-v `pwd`:/src \
		-w /src/bin \
		bearstech/upx \
		upx redistop
