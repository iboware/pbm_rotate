.PHONY: test
test.unit:
	echo "=> Running Tests"
	go test -v ./...

.PHONY: build
build:
	echo "=> Building..."
	CGO_ENABLED=0 go build -a -ldflags '-w -s' -o bin/pbm_rotate

.PHONY: run
run:
	./bin/pbm_rotate