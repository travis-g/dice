fmt:
	@echo "==> Running gofmt..."
	gofmt -s -w .

build: fmt test
	@echo "==> Building library..."
	go build -ldflags="-s -w" ./...
	@echo "==> Building the CLI..."
	go build -ldflags="-s -w" ./cmd/dice

test:
	@echo "==> Running tests..."
	@go test -cover ./...

report:
	@echo "==> Generating report card..."
	@goreportcard-cli -v

bench: test
	@echo "==> Running benchmarks (may take a while)..."
	@go test -run=XXX -bench=. ./...

cover:
	@echo "==> Calculating coverage..."
	@go test -coverprofile=coverage.out . ./math
	@go tool cover -func=coverage.out | grep -vE "^total" | sort -k3,3n
	@go tool cover -html=coverage.out

clean:
	@rm -f dice dice.exe parser parser.exe coverage.out

godoc:
	@echo "==> View godoc at http://localhost:8080/pkg/github.com/travis-g/dice/"
	@godoc -http ":8080"

.PHONY: proto
proto:
	protoc ./*.proto --go_out=plugins=grpc,paths=source_relative:.

.PHONY: clean build godoc
