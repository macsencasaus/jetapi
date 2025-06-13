build:
	go build -o ./bin/jetapi ./cmd/jetapi
.PHONY: build

run: build
	./bin/jetapi
.PHONY: run

run-dev:
	air --build.cmd "go build -o ./bin/jetapi ./cmd/jetapi" --build.bin "./bin/jetapi"
.PHONY: run-dev

clean:
	rm -r bin/
	rm -r tmp/
	rm -r cmd/jetapi/tmp
.PHONY: clean
