build:
	go build .

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
		-w /src \
		bearstech/upx \
		upx redistop
