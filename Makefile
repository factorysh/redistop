build:
	go build .

test:
	go test --cover \
		github.com/factorysh/redistop/stats
