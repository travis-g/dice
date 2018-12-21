build:
	@echo "==> Building the CLI..."
	go build -ldflags="-s -w" ../dice/cmd/dice

clean:
	@rm dice dice.exe

.PHONY: clean build
