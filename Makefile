build: test
	@echo "==> Building the CLI..."
	go build -ldflags="-s -w" ./cmd/dice

test:
	@echo "==> Running tests..."
	@go test -cover . ./math

bench: test
	@echo "==> Running benchmarks (may take a while)..."
	@go test -run=XXX -bench=. -benchtime=5s . ./math

cover:
	@echo "==> Calculating coverage..."
	@go test -coverprofile=coverage.out . ./math
	@go tool cover -func=coverage.out | grep -vE "^total" | sort -k3,3n
	@go tool cover -html=coverage.out

clean:
	@rm -f dice dice.exe coverage.out

.PHONY: clean build
